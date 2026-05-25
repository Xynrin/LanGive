package com.langive.app.ui

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.langive.app.service.LanGiveService
import com.langive.app.service.TransferItem

@Composable
fun TransfersScreen(service: LanGiveService?, modifier: Modifier = Modifier) {
    val transfers by if (service != null) service.transfers.collectAsState() else remember { mutableStateOf(emptyList()) }

    Column(
        modifier = modifier
            .fillMaxSize()
            .padding(16.dp)
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Column {
                Text("传输记录", fontSize = 20.sp, style = MaterialTheme.typography.titleLarge)
                Text(
                    "全部历史传输",
                    fontSize = 12.sp,
                    color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.6f)
                )
            }
            Button(
                onClick = { service?.clearCompletedTransfers() }
            ) {
                Text("清除已完成")
            }
        }

        Spacer(modifier = Modifier.height(16.dp))

        if (transfers.isEmpty()) {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text("📭", fontSize = 48.sp)
                    Spacer(modifier = Modifier.height(10.dp))
                    Text("暂无传输记录", color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.6f))
                }
            }
        } else {
            LazyColumn(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                items(transfers) { transfer ->
                    TransferCard(transfer = transfer, onCancel = { service?.cancelTransfer(transfer.id) })
                }
            }
        }
    }
}

@Composable
fun TransferCard(transfer: TransferItem, onCancel: () -> Unit) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text(if (transfer.type == "send") "📤" else "📥", fontSize = 20.sp, modifier = Modifier.padding(end = 8.dp))
                    Text(
                        transfer.fileName,
                        fontSize = 14.sp,
                        fontWeight = FontWeight.Bold,
                        maxLines = 1,
                        modifier = Modifier.widthIn(max = 200.dp)
                    )
                }

                val statusText = when (transfer.status) {
                    "pending" -> "等待中"
                    "transferring" -> "传输中"
                    "completed" -> "已完成"
                    "failed" -> "失败"
                    "cancelled" -> "已取消"
                    else -> transfer.status
                }
                val statusColor = when (transfer.status) {
                    "completed" -> MaterialTheme.colorScheme.secondary
                    "transferring" -> MaterialTheme.colorScheme.primary
                    "failed" -> MaterialTheme.colorScheme.error
                    else -> MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
                }

                Text(statusText, fontSize = 12.sp, color = statusColor, fontWeight = FontWeight.Bold)
            }

            Spacer(modifier = Modifier.height(10.dp))

            if (transfer.status == "transferring") {
                LinearProgressIndicator(
                    progress = (transfer.progress / 100f).toFloat(),
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(6.dp)
                )
                Spacer(modifier = Modifier.height(6.dp))
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween
                ) {
                    Text(
                        "${formatSize(transfer.sentSize)} / ${formatSize(transfer.totalSize)}",
                        fontSize = 11.sp,
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
                    )
                    Text(
                        "${String.format("%.1f", transfer.progress)}%",
                        fontSize = 11.sp,
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
                    )
                }
            } else {
                Text(
                    "大小: ${formatSize(transfer.totalSize)}",
                    fontSize = 12.sp,
                    color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
                )
            }

            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = 8.dp),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    "对端: ${transfer.peerAddr}",
                    fontSize = 11.sp,
                    color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.4f)
                )
                if (transfer.status == "transferring") {
                    Button(
                        onClick = onCancel,
                        colors = ButtonDefaults.buttonColors(containerColor = MaterialTheme.colorScheme.error),
                        contentPadding = PaddingValues(horizontal = 12.dp, vertical = 4.dp),
                        modifier = Modifier.height(28.dp)
                    ) {
                        Text("取消", fontSize = 11.sp)
                    }
                }
            }
        }
    }
}
