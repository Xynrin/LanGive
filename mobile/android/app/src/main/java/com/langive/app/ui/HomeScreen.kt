package com.langive.app.ui

import android.content.Intent
import android.net.Uri
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.langive.app.R
import com.langive.app.service.LanGiveService

@Composable
fun HomeScreen(service: LanGiveService?, modifier: Modifier = Modifier) {
    val context = LocalContext.current
    val devices by if (service != null) service.devices.collectAsState() else remember { mutableStateOf(emptyList()) }
    val transfers by if (service != null) service.transfers.collectAsState() else remember { mutableStateOf(emptyList()) }

    val completedCount = transfers.count { it.status == "completed" }
    val activeCount = transfers.count { it.status == "transferring" }

    // SAF directory picker launcher for "Open Download Folder"
    val openDirLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.OpenDocumentTree()
    ) { _ ->
        // No action needed as SAF handles tree browsing
    }

    Column(
        modifier = modifier
            .fillMaxSize()
            .padding(20.dp)
            .verticalScroll(rememberScrollState()),
        verticalArrangement = Arrangement.Top,
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Spacer(modifier = Modifier.height(20.dp))
        Text(
            text = stringResource(R.string.welcome_title),
            fontSize = 24.sp,
            fontWeight = FontWeight.Bold,
            color = MaterialTheme.colorScheme.primary
        )
        Text(
            text = stringResource(R.string.welcome_subtitle),
            fontSize = 14.sp,
            color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.7f),
            modifier = Modifier.padding(top = 4.dp)
        )

        Spacer(modifier = Modifier.height(40.dp))

        // Stats grid
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            StatCard(
                value = devices.size.toString(),
                label = stringResource(R.string.stat_online_devices),
                modifier = Modifier.weight(1f)
            )
            StatCard(
                value = completedCount.toString(),
                label = stringResource(R.string.stat_completed_transfers),
                modifier = Modifier.weight(1f)
            )
            StatCard(
                value = activeCount.toString(),
                label = stringResource(R.string.stat_active_transfers),
                modifier = Modifier.weight(1f)
            )
        }

        Spacer(modifier = Modifier.height(40.dp))

        // Actions
        Button(
            onClick = {
                service?.downloadPathState?.value?.let { path ->
                    val intent = Intent(Intent.ACTION_OPEN_DOCUMENT_TREE)
                    context.startActivity(Intent.createChooser(intent, "打开下载目录"))
                } ?: run {
                    openDirLauncher.launch(null)
                }
            },
            modifier = Modifier
                .fillMaxWidth()
                .height(56.dp)
        ) {
            Text(stringResource(R.string.action_open_downloads), fontSize = 16.sp)
        }
    }
}

@Composable
fun StatCard(value: String, label: String, modifier: Modifier = Modifier) {
    Card(
        modifier = modifier,
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface)
    ) {
        Column(
            modifier = Modifier
                .padding(16.dp)
                .fillMaxWidth(),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            Text(
                text = value,
                fontSize = 28.sp,
                fontWeight = FontWeight.Bold,
                color = MaterialTheme.colorScheme.primary
            )
            Spacer(modifier = Modifier.height(4.dp))
            Text(
                text = label,
                fontSize = 12.sp,
                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
            )
        }
    }
}
