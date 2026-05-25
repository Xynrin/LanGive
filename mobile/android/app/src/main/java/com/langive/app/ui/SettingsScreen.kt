package com.langive.app.ui

import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.langive.app.service.LanGiveService

@Composable
fun SettingsScreen(service: LanGiveService?, modifier: Modifier = Modifier) {
    if (service == null) {
        Box(modifier = modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
            CircularProgressIndicator()
        }
        return
    }

    val deviceName by service.deviceNameState.collectAsState()
    val downloadPath by service.downloadPathState.collectAsState()
    val privacyMode by service.privacyModeState.collectAsState()
    val port by service.portState.collectAsState()
    val scanInterval by service.scanIntervalState.collectAsState()
    val sessionID by service.sessionIDState.collectAsState()

    var nameInput by remember { mutableStateOf(deviceName) }
    var portInput by remember { mutableStateOf(port.toString()) }

    // Sync input variables if model changes
    LaunchedEffect(deviceName) { nameInput = deviceName }
    LaunchedEffect(port) { portInput = port.toString() }

    val openDirLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.OpenDocumentTree()
    ) { uri ->
        if (uri != null) {
            // Note: In real app, we convert Uri to actual path or persist permission
            service.updateDownloadPath(uri.path ?: "")
        }
    }

    Column(
        modifier = modifier
            .fillMaxSize()
            .padding(16.dp)
            .verticalScroll(rememberScrollState()),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        Text("设置", fontSize = 20.sp, style = MaterialTheme.typography.titleLarge)

        // Device Info
        Card(
            modifier = Modifier.fillMaxWidth(),
            colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface)
        ) {
            Column(modifier = Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
                Text("设备信息", fontWeight = FontWeight.Bold, fontSize = 14.sp)
                
                OutlinedTextField(
                    value = nameInput,
                    onValueChange = { nameInput = it },
                    label = { Text("设备名称") },
                    trailingIcon = {
                        TextButton(onClick = { service.updateDeviceName(nameInput) }) {
                            Text("保存")
                        }
                    },
                    modifier = Modifier.fillMaxWidth()
                )

                Text(
                    "此名称将显示在其他设备的发现列表中",
                    fontSize = 11.sp,
                    color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f)
                )
            }
        }

        // Privacy & Session
        Card(
            modifier = Modifier.fillMaxWidth(),
            colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface)
        ) {
            Column(modifier = Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
                Text("隐私与会话", fontWeight = FontWeight.Bold, fontSize = 14.sp)

                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Column(modifier = Modifier.weight(1f)) {
                        Text("隐私模式", fontSize = 15.sp)
                        Text(
                            "开启后不在公共会话中广播，仅能通过 IP 直接连接",
                            fontSize = 11.sp,
                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f)
                        )
                    }
                    Switch(
                        checked = privacyMode,
                        onCheckedChange = { service.updatePrivacyMode(it) }
                    )
                }

                if (!privacyMode) {
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Text("当前会话", fontSize = 14.sp)
                        AssistChip(onClick = {}, label = { Text("公共会话") })
                    }
                } else {
                    Column {
                        Text("会话 ID", fontSize = 12.sp, color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f))
                        Text(sessionID, fontSize = 12.sp, style = MaterialTheme.typography.bodySmall)
                    }
                }
            }
        }

        // Storage Setting
        Card(
            modifier = Modifier.fillMaxWidth(),
            colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface)
        ) {
            Column(modifier = Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
                Text("存储设置", fontWeight = FontWeight.Bold, fontSize = 14.sp)
                
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Column(modifier = Modifier.weight(1f)) {
                        Text("下载路径", fontSize = 14.sp)
                        Text(
                            downloadPath,
                            fontSize = 11.sp,
                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f)
                        )
                    }
                    Button(onClick = { openDirLauncher.launch(null) }) {
                        Text("更改")
                    }
                }
            }
        }

        // Scan Setting
        Card(
            modifier = Modifier.fillMaxWidth(),
            colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface)
        ) {
            Column(modifier = Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
                Text("扫描设置", fontWeight = FontWeight.Bold, fontSize = 14.sp)
                Text("扫描间隔", fontSize = 14.sp)
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    val intervals = listOf(5, 15, 30)
                    val labels = listOf("快速 (5秒)", "平衡 (15秒)", "节能 (30秒)")
                    intervals.forEachIndexed { i, intervalSecs ->
                        val active = scanInterval == intervalSecs
                        val containerColor = if (active) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.surfaceVariant
                        val contentColor = if (active) MaterialTheme.colorScheme.onPrimary else MaterialTheme.colorScheme.onSurfaceVariant
                        
                        Button(
                            onClick = { service.updateScanInterval(intervalSecs) },
                            colors = ButtonDefaults.buttonColors(containerColor = containerColor, contentColor = contentColor),
                            contentPadding = PaddingValues(horizontal = 6.dp),
                            modifier = Modifier.weight(1f)
                        ) {
                            Text(labels[i], fontSize = 11.sp)
                        }
                    }
                }
                Text(
                    "后台运行时将自动使用更长的扫描间隔",
                    fontSize = 11.sp,
                    color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f)
                )
            }
        }

        // Advanced Setting
        Card(
            modifier = Modifier.fillMaxWidth(),
            colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface)
        ) {
            Column(modifier = Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
                Text("高级设置", fontWeight = FontWeight.Bold, fontSize = 14.sp)
                
                OutlinedTextField(
                    value = portInput,
                    onValueChange = { portInput = it },
                    label = { Text("服务端口") },
                    keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
                    trailingIcon = {
                        TextButton(
                            onClick = {
                                portInput.toIntOrNull()?.let {
                                    if (it in 1024..65535) {
                                        service.updatePort(it)
                                    }
                                }
                            }
                        ) {
                            Text("保存")
                        }
                    },
                    modifier = Modifier.fillMaxWidth()
                )
                Text(
                    "注意：修改端口后需要重启应用服务方可生效",
                    fontSize = 11.sp,
                    color = MaterialTheme.colorScheme.error
                )
            }
        }

        // About & Updates
        Card(
            modifier = Modifier.fillMaxWidth(),
            colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface)
        ) {
            Column(modifier = Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                Text("关于与更新", fontWeight = FontWeight.Bold, fontSize = 14.sp)
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween
                ) {
                    Text("当前版本", fontSize = 14.sp)
                    Text("v1.0.0", fontSize = 14.sp)
                }
                Divider(modifier = Modifier.padding(vertical = 4.dp))
                Text(
                    "Android 不支持自动更新，请到 GitHub Releases 下载新版",
                    fontSize = 12.sp,
                    color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
                )
            }
        }
    }
}
