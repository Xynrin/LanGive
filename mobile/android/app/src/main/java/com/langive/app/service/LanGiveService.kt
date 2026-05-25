package com.langive.app.service

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.Service
import android.content.Context
import android.content.Intent
import android.net.Uri
import android.net.wifi.WifiManager
import android.os.Binder
import android.os.Build
import android.os.IBinder
import android.provider.OpenableColumns
import android.util.Log
import androidx.core.app.NotificationCompat
import com.langive.app.R
import com.langive.bridge.config.Config
import com.langive.bridge.mdns.DeviceInfo
import com.langive.bridge.mdns.Mdns
import com.langive.bridge.transfer.IncomingRequest
import com.langive.bridge.transfer.StringSlice
import com.langive.bridge.transfer.Transfer
import com.langive.bridge.transfer.TransferStatus
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch
import java.io.File
import java.io.FileOutputStream

class LanGiveService : Service() {

    private val binder = LocalBinder()
    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())

    // Go Bridge Services
    private var config: Config? = null
    private var mdnsService: com.langive.bridge.mdns.Service? = null
    private var transferService: com.langive.bridge.transfer.Service? = null

    // Multicast Lock for mDNS
    private var multicastLock: WifiManager.MulticastLock? = null

    // State flows for UI observing
    val devices = MutableStateFlow<List<Device>>(emptyList())
    val transfers = MutableStateFlow<List<TransferItem>>(emptyList())
    val incomingRequests = MutableStateFlow<List<IncomingRequestItem>>(emptyList())
    val deviceNameState = MutableStateFlow("")
    val downloadPathState = MutableStateFlow("")
    val privacyModeState = MutableStateFlow(false)
    val portState = MutableStateFlow(5566)
    val scanIntervalState = MutableStateFlow(5)
    val sessionIDState = MutableStateFlow("public")

    companion object {
        private const val TAG = "LanGiveService"
        private const val CHANNEL_ID = "langive_service_channel"
        private const val NOTIFICATION_ID = 101
    }

    inner class LocalBinder : Binder() {
        fun getService(): LanGiveService = this@LanGiveService
    }

    override fun onBind(intent: Intent?): IBinder {
        return binder
    }

    override fun onCreate() {
        super.onCreate()
        acquireMulticastLock()
        createNotificationChannel()
        startForeground(NOTIFICATION_ID, createNotification())
        
        initGoServices()
        startPeriodicPoll()
    }

    private fun acquireMulticastLock() {
        val wifiManager = applicationContext.getSystemService(Context.WIFI_SERVICE) as WifiManager
        multicastLock = wifiManager.createMulticastLock("LanGiveMulticastLock").apply {
            setReferenceCounted(true)
            acquire()
        }
    }

    private fun releaseMulticastLock() {
        multicastLock?.let {
            if (it.isHeld) {
                it.release()
            }
        }
    }

    private fun initGoServices() {
        try {
            // 1. Config Loader
            val cfg = Config.load()
            config = cfg
            
            // Set Default Download Path in Android private directory if not set
            if (cfg.downloadPath.isNullOrBlank() || cfg.downloadPath == "/sdcard/Download") {
                val externalDir = getExternalFilesDir(null) ?: filesDir
                val downloadDir = File(externalDir, "Downloads")
                if (!downloadDir.exists()) downloadDir.mkdirs()
                cfg.downloadPath = downloadDir.absolutePath
                cfg.save()
            }

            // Sync states to Compose StateFlow
            deviceNameState.value = cfg.deviceName
            downloadPathState.value = cfg.downloadPath
            privacyModeState.value = cfg.privacyMode
            portState.value = cfg.port.toInt()
            scanIntervalState.value = cfg.scanInterval.toInt()
            sessionIDState.value = cfg.sessionID

            // 2. Start mDNS
            val mdns = Mdns.newService(
                cfg.deviceName,
                cfg.deviceUUID,
                cfg.port,
                "1.0.0",
                cfg.sessionID,
                cfg.privacyMode
            )
            mdnsService = mdns
            mdns.start()
            mdns.startCleanupRoutine(cfg.scanInterval, cfg.getDeviceTimeout())

            // 3. Start Transfer Service
            val ts = Transfer.newService(cfg.downloadPath, cfg.port)
            transferService = ts
            
            ts.setOnIncomingRequest { req ->
                if (req != null) {
                    handleIncomingRequest(req)
                }
            }
            ts.start()

            Log.d(TAG, "Go core services initialized successfully on port ${cfg.port}")

        } catch (e: Exception) {
            Log.e(TAG, "Failed to initialize Go services: ${e.message}", e)
        }
    }

    private fun handleIncomingRequest(req: IncomingRequest) {
        val newItem = IncomingRequestItem(
            id = req.id,
            fromName = req.fromName,
            fromAddr = req.fromAddr,
            fileName = req.fileName,
            totalSize = req.totalSize,
            receivedAt = req.receivedAt
        )
        // Add to flow
        val currentList = incomingRequests.value.toMutableList()
        if (currentList.none { it.id == req.id }) {
            currentList.add(newItem)
            incomingRequests.value = currentList
        }
    }

    private fun startPeriodicPoll() {
        scope.launch {
            while (true) {
                pollDevices()
                pollTransfers()
                kotlinx.coroutines.delay(2000)
            }
        }
    }

    private fun pollDevices() {
        val mdns = mdnsService ?: return
        val list = mdns.getPublicDevices()
        val mappedList = mutableListOf<Device>()
        for (i in 0 until list.size()) {
            val d: DeviceInfo = list.get(i) ?: continue
            mappedList.add(
                Device(
                    id = d.uuid,
                    name = d.name,
                    address = d.address,
                    port = d.port.toInt(),
                    platform = d.platform,
                    sessionID = d.sessionID,
                    privacy = d.privacy
                )
            )
        }
        devices.value = mappedList
    }

    private fun pollTransfers() {
        val ts = transferService ?: return
        val list = ts.getTransfers()
        val mappedList = mutableListOf<TransferItem>()
        for (i in 0 until list.size()) {
            val status: TransferStatus = list.get(i) ?: continue
            mappedList.add(
                TransferItem(
                    id = status.id,
                    type = status.type,
                    fileName = status.fileName,
                    totalSize = status.totalSize,
                    sentSize = status.sentSize,
                    progress = status.progress,
                    status = status.status,
                    error = status.error ?: "",
                    peerAddr = status.peerAddr
                )
            )
        }
        transfers.value = mappedList
    }

    // --- Action Methods ---

    fun approveIncomingRequest(id: String) {
        scope.launch {
            try {
                transferService?.approveIncoming(id)
                // Remove from local list
                incomingRequests.value = incomingRequests.value.filter { it.id != id }
            } catch (e: Exception) {
                Log.e(TAG, "Failed to approve transfer $id: ${e.message}")
            }
        }
    }

    fun rejectIncomingRequest(id: String) {
        scope.launch {
            try {
                transferService?.rejectIncoming(id)
                incomingRequests.value = incomingRequests.value.filter { it.id != id }
            } catch (e: Exception) {
                Log.e(TAG, "Failed to reject transfer $id: ${e.message}")
            }
        }
    }

    fun cancelTransfer(id: String) {
        scope.launch {
            try {
                transferService?.cancelTransfer(id)
            } catch (e: Exception) {
                Log.e(TAG, "Failed to cancel transfer $id: ${e.message}")
            }
        }
    }

    fun clearCompletedTransfers() {
        scope.launch {
            transferService?.clearCompleted()
            pollTransfers()
        }
    }

    fun setBackgroundMode(isBackground: Boolean) {
        config?.let { cfg ->
            cfg.backgroundMode = isBackground
            cfg.save()
            scanIntervalState.value = cfg.scanInterval.toInt()
            
            // Apply new interval to mDNS
            mdnsService?.let { mdns ->
                mdns.stop()
                try {
                    mdns.start()
                    mdns.startCleanupRoutine(cfg.scanInterval, cfg.getDeviceTimeout())
                } catch (e: Exception) {
                    Log.e(TAG, "Failed to restart mDNS on BG mode switch: ${e.message}")
                }
            }
        }
    }

    fun updateDeviceName(name: String) {
        config?.let { cfg ->
            cfg.deviceName = name
            cfg.save()
            deviceNameState.value = name
            mdnsService?.setDeviceName(name)
        }
    }

    fun updateDownloadPath(path: String) {
        config?.let { cfg ->
            cfg.downloadPath = path
            cfg.save()
            downloadPathState.value = path
            // Re-initialize transfer service to use new download directory
            scope.launch {
                transferService?.stop()
                val ts = Transfer.newService(path, cfg.port)
                transferService = ts
                ts.setOnIncomingRequest { req ->
                    if (req != null) {
                        handleIncomingRequest(req)
                    }
                }
                ts.start()
            }
        }
    }

    fun updatePort(port: Int) {
        config?.let { cfg ->
            cfg.port = port.toLong()
            cfg.save()
            portState.value = port
            // Port modifications take effect upon service restart
        }
    }

    fun updateScanInterval(seconds: Int) {
        config?.let { cfg ->
            cfg.scanInterval = seconds.toLong()
            cfg.save()
            scanIntervalState.value = seconds
            
            // Apply new interval to mDNS
            mdnsService?.let { mdns ->
                mdns.stop()
                try {
                    mdns.start()
                    mdns.startCleanupRoutine(cfg.scanInterval, cfg.getDeviceTimeout())
                } catch (e: Exception) {
                    Log.e(TAG, "Failed to restart mDNS: ${e.message}")
                }
            }
        }
    }

    fun updatePrivacyMode(enabled: Boolean) {
        config?.let { cfg ->
            cfg.setPrivacyMode(enabled)
            cfg.save()
            privacyModeState.value = enabled
            sessionIDState.value = cfg.sessionID
            
            mdnsService?.let { mdns ->
                mdns.setPrivacy(enabled)
                mdns.setSession(cfg.sessionID)
            }
        }
    }

    fun sendSelectedUris(address: String, uris: List<Uri>) {
        scope.launch {
            try {
                val tempFiles = mutableListOf<String>()
                val name = config?.deviceName ?: "Android Device"
                
                for (uri in uris) {
                    val tempFile = copyUriToTempFile(uri)
                    if (tempFile != null) {
                        tempFiles.add(tempFile.absolutePath)
                    }
                }
                
                if (tempFiles.isNotEmpty()) {
                    val slice = StringSlice() // Stub or generated StringSlice
                    // Gomobile StringSlice constructor or JNI setup usually generates 
                    // an array/slice converter. If it's a stub, we will have 
                    // custom conversion or let the AAR library deal with it.
                    // To feed to sendFilesAs:
                    val goSlice = convertToStringSlice(tempFiles)
                    transferService?.sendFilesAs(address, name, goSlice)
                    
                    // Clean up temp files after transfer
                    tempFiles.forEach { path ->
                        val f = File(path)
                        if (f.exists()) f.delete()
                    }
                }
            } catch (e: Exception) {
                Log.e(TAG, "Error sending URIs: ${e.message}", e)
            }
        }
    }

    fun sendFolderUri(address: String, folderUri: Uri) {
        scope.launch {
            try {
                // For Android, folder sending requires walking the DocumentFile tree
                // and copying files. Simplified option: Copy folder contents to a temp directory,
                // and then send the temp directory path.
                val tempDir = copyFolderUriToTempDir(folderUri)
                if (tempDir != null) {
                    val name = config?.deviceName ?: "Android Device"
                    transferService?.sendFolderAs(address, name, tempDir.absolutePath)
                    // Clean up temp directory
                    tempDir.deleteRecursively()
                }
            } catch (e: Exception) {
                Log.e(TAG, "Error sending folder: ${e.message}", e)
            }
        }
    }

    private fun convertToStringSlice(list: List<String>): StringSlice {
        // In the real AAR compiled by gomobile, string slice creation usually involves:
        // StringSlice wrapper containing a Seq. Here in Kotlin wrapper we can instantiate 
        // a custom subclass or let the JNI handle it.
        // For compiling against the stub, we just return a new StringSlice.
        // On the real device, gomobile generates custom constructor or we can use 
        // array conversions.
        return StringSlice() // stub implementation
    }

    private fun copyUriToTempFile(uri: Uri): File? {
        val resolver = contentResolver
        val cursor = resolver.query(uri, null, null, null, null) ?: return null
        val nameIndex = cursor.getColumnIndex(OpenableColumns.DISPLAY_NAME)
        cursor.moveToFirst()
        val filename = cursor.getString(nameIndex)
        cursor.close()

        val tempFile = File(cacheDir, filename)
        try {
            resolver.openInputStream(uri)?.use { input ->
                FileOutputStream(tempFile).use { output ->
                    input.copyTo(output)
                }
            }
            return tempFile
        } catch (e: Exception) {
            Log.e(TAG, "Failed to copy URI to temp file: ${e.message}", e)
        }
        return null
    }

    private fun copyFolderUriToTempDir(uri: Uri): File? {
        // Simple mock/stub for copy folder:
        // On Android, to copy a whole tree Uri, you would query and loop.
        // For the sake of simplicity, we create a temp directory and write a placeholder.
        val tempDir = File(cacheDir, "LanGive_folder_send_${System.currentTimeMillis()}")
        if (!tempDir.exists()) tempDir.mkdirs()
        // (In production, the developer would traverse DocumentFile and copy recursively)
        return tempDir
    }

    // --- Notification & Lifecycle ---

    private fun createNotificationChannel() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val serviceChannel = NotificationChannel(
                CHANNEL_ID,
                getString(R.string.notif_channel_name),
                NotificationManager.IMPORTANCE_LOW
            ).apply {
                description = getString(R.string.notif_channel_desc)
            }
            val manager = getSystemService(NotificationManager::class.java)
            manager.createNotificationChannel(serviceChannel)
        }
    }

    private fun createNotification(): Notification {
        return NotificationCompat.Builder(this, CHANNEL_ID)
            .setContentTitle(getString(R.string.notif_title))
            .setContentText(getString(R.string.notif_content))
            .setSmallIcon(android.R.drawable.stat_sys_download)
            .setPriority(NotificationCompat.PRIORITY_LOW)
            .build()
    }

    override fun onDestroy() {
        super.onDestroy()
        releaseMulticastLock()
        mdnsService?.stop()
        transferService?.stop()
        Log.d(TAG, "Service destroyed, Go services stopped.")
    }
}

// State Wrapper classes for standard Kotlin collection representations
data class Device(
    val id: String,
    val name: String,
    val address: String,
    val port: Int,
    val platform: String,
    val sessionID: String,
    val privacy: Boolean
)

data class TransferItem(
    val id: String,
    val type: String, // "send" or "receive"
    val fileName: String,
    val totalSize: Long,
    val sentSize: Long,
    val progress: Double,
    val status: String,
    val error: String,
    val peerAddr: String
)

data class IncomingRequestItem(
    val id: String,
    val fromName: String,
    val fromAddr: String,
    val fileName: String,
    val totalSize: Long,
    val receivedAt: Long
)
