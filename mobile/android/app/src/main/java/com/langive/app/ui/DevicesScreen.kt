package com.langive.app.ui

import android.net.Uri
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.langive.app.service.Device
import com.langive.app.service.LanGiveService

@Composable
fun DevicesScreen(service: LanGiveService?, modifier: Modifier = Modifier) {
    val devices by if (service != null) service.devices.collectAsState() else remember { mutableStateOf(emptyList()) }

    var selectedDevice by remember { mutableStateOf<Device?>(null) }
    var selectedFiles by remember { mutableStateOf<List<Uri>>(emptyList()) }
    var selectedFolder by remember { mutableStateOf<Uri?>(null) }
    var showSendDialog by remember { mutableStateOf(false) }

    val filePicker = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.OpenMultipleDocuments()
    ) { uris ->
        if (uris.isNotEmpty()) {
            selectedFiles = uris
            selectedFolder = null
        }
    }

    val folderPicker = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.OpenDocumentTree()
    ) { uri ->
        if (uri != null) {
            selectedFolder = uri
            selectedFiles = emptyList()
        }
    }

    Column(
        modifier = modifier
            .fillMaxSize()
            .padding(16.dp)
    ) {
        Text("在线设备", fontSize = 20.sp, style = MaterialTheme.typography.titleLarge)
        Text(
            "发现 ${devices.size} 台局域网设备",
            fontSize = 14.sp,
            color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.6f),
            modifier = Modifier.padding(bottom = 16.dp)
        )

        if (devices.isEmpty()) {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text("🔍", fontSize = 48.sp)
                    Spacer(modifier = Modifier.height(10.dp))
                    Text("未发现在线设备", color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.6f))
                }
            }
        } else {
            LazyColumn(
                verticalArrangement = Arrangement.spacedBy(8.dp),
                modifier = Modifier.weight(1f)
            ) {
                items(devices) { device ->
                    DeviceCard(
                        device = device,
                        onClick = {
                            selectedDevice = device
                            selectedFiles = emptyList()
                            selectedFolder = null
                            showSendDialog = true
                        }
                    )
                }
            }
        }

        if (showSendDialog && selectedDevice != null) {
            AlertDialog(
                onDismissRequest = { showSendDialog = false },
                title = { Text("发送文件到 ${selectedDevice?.name}") },
                text = {
                    Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                        Row(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                            Button(onClick = { filePicker.launch(arrayOf("*/*")) }) {
                                Text("选择文件")
                            }
                            Button(onClick = { folderPicker.launch(null) }) {
                                Text("选择文件夹")
                            }
                        }

                        if (selectedFiles.isNotEmpty()) {
                            Text("已选择 ${selectedFiles.size} 个文件")
                        } else if (selectedFolder != null) {
                            Text("已选择文件夹: ${selectedFolder?.lastPathSegment}")
                        }
                    }
                },
                confirmButton = {
                    Button(
                        onClick = {
                            val dev = selectedDevice
                            if (dev != null) {
                                if (selectedFiles.isNotEmpty()) {
                                    service?.sendSelectedUris(dev.address, selectedFiles)
                                } else if (selectedFolder != null) {
                                    service?.sendFolderUri(dev.address, selectedFolder!!)
                                }
                            }
                            showSendDialog = false
                        },
                        enabled = selectedFiles.isNotEmpty() || selectedFolder != null
                    ) {
                        Text("发送")
                    }
                },
                dismissButton = {
                    TextButton(onClick = { showSendDialog = false }) {
                        Text("取消")
                    }
                }
            )
        }
    }
}

@Composable
fun DeviceCard(device: Device, onClick: () -> Unit) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .clickable { onClick() },
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface)
    ) {
        Row(
            modifier = Modifier.padding(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            val icon = when (device.platform.lowercase()) {
                "windows" -> "💻"
                "darwin", "macos" -> "🍎"
                "linux" -> "🐧"
                "android", "ios" -> "📱"
                else -> "💻"
            }
            Text(icon, fontSize = 32.sp, modifier = Modifier.padding(end = 16.dp))
            Column(modifier = Modifier.weight(1f)) {
                Text(device.name, fontSize = 16.sp, style = MaterialTheme.typography.titleMedium)
                Text(
                    device.address,
                    fontSize = 12.sp,
                    color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
                )
            }
            Button(onClick = onClick) {
                Text("发送")
            }
        }
    }
}
