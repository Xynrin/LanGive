package com.langive.app.ui

import androidx.compose.foundation.layout.padding
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import com.langive.app.service.LanGiveService
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Home
import androidx.compose.material.icons.filled.Info
import androidx.compose.material.icons.filled.List
import androidx.compose.material.icons.filled.Settings
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.res.stringResource
import com.langive.app.R

@Composable
fun MainScreen(service: LanGiveService?) {
    var selectedTab by remember { mutableIntStateOf(0) }
    
    val incomingRequests by if (service != null) {
        service.incomingRequests.collectAsState()
    } else {
        remember { mutableStateOf(emptyList()) }
    }

    Scaffold(
        bottomBar = {
            NavigationBar {
                val tabs = listOf(
                    TabItem(stringResource(R.string.title_home), Icons.Default.Home),
                    TabItem(stringResource(R.string.title_devices), Icons.Default.Info),
                    TabItem(stringResource(R.string.title_transfers), Icons.Default.List),
                    TabItem(stringResource(R.string.title_settings), Icons.Default.Settings)
                )
                tabs.forEachIndexed { index, tab ->
                    NavigationBarItem(
                        icon = { Icon(tab.icon, contentDescription = tab.title) },
                        label = { Text(tab.title) },
                        selected = selectedTab == index,
                        onClick = { selectedTab = index }
                    )
                }
            }
        }
    ) { innerPadding ->
        Modifier.padding(innerPadding)
        val modifier = Modifier.padding(innerPadding)
        
        when (selectedTab) {
            0 -> HomeScreen(service, modifier)
            1 -> DevicesScreen(service, modifier)
            2 -> TransfersScreen(service, modifier)
            3 -> SettingsScreen(service, modifier)
        }

        // Show pop-up dialog for incoming file requests
        if (incomingRequests.isNotEmpty() && service != null) {
            val req = incomingRequests.first()
            AlertDialog(
                onDismissRequest = { /* Force explicit accept or reject */ },
                title = { Text("收到传输请求") },
                text = {
                    Text(
                        "来自设备 ${req.fromName} (${req.fromAddr}) 申请发送文件：\n\n" +
                        "文件名: ${req.fileName}\n" +
                        "大小: ${formatSize(req.totalSize)}"
                    )
                },
                confirmButton = {
                    Button(
                        onClick = { service.approveIncomingRequest(req.id) }
                    ) {
                        Text("接收")
                    }
                },
                dismissButton = {
                    TextButton(
                        onClick = { service.rejectIncomingRequest(req.id) }
                    ) {
                        Text("拒绝")
                    }
                }
            )
        }
    }
}

private data class TabItem(val title: String, val icon: ImageVector)

fun formatSize(bytes: Long): String {
    if (bytes <= 0) return "0 B"
    val k = 1024.0
    val sizes = arrayOf("B", "KB", "MB", "GB", "TB")
    val i = kotlin.math.floor(kotlin.math.log(bytes.toDouble(), k)).toInt()
    return String.format("%.2f %s", bytes / Math.pow(k, i.toDouble()), sizes[i])
}
