package org.trainblink.app.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.Info
import androidx.compose.material.icons.filled.Send
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import kotlinx.coroutines.launch
import org.trainblink.app.network.MatrixEphemeralRoom
import org.trainblink.app.services.MatrixChatService
import org.trainblink.app.services.MessageItem
import java.text.SimpleDateFormat
import java.util.*

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ChatScreen(
    ephemeralRoom: MatrixEphemeralRoom,
    chatService: MatrixChatService,
    onBack: () -> Unit
) {
    var messageText by remember { mutableStateOf("") }
    var showInfo by remember { mutableStateOf(false) }
    var showExtendLifetime by remember { mutableStateOf(false) }

    val messages by chatService.messages.collectAsState()
    val listState = rememberLazyListState()
    val scope = rememberCoroutineScope()

    // Open room on first composition
    LaunchedEffect(ephemeralRoom.roomId) {
        try {
            chatService.openRoom(ephemeralRoom)
        } catch (e: Exception) {
            // Handle error
        }
    }

    // Scroll to bottom when new messages arrive
    LaunchedEffect(messages.size) {
        if (messages.isNotEmpty()) {
            listState.animateScrollToItem(messages.size - 1)
        }
    }

    // Cleanup on dispose
    DisposableEffect(Unit) {
        onDispose {
            chatService.closeRoom()
        }
    }

    Scaffold(
        topBar = {
            ChatTopBar(
                room = ephemeralRoom,
                onBack = onBack,
                onInfo = { showInfo = true },
                onExtend = { showExtendLifetime = true }
            )
        },
        bottomBar = {
            MessageInputBar(
                text = messageText,
                onTextChange = { messageText = it },
                onSend = {
                    scope.launch {
                        try {
                            chatService.sendMessage(messageText)
                            messageText = ""
                        } catch (e: Exception) {
                            // Handle error
                        }
                    }
                }
            )
        }
    ) { paddingValues ->
        LazyColumn(
            modifier = Modifier
                .fillMaxSize()
                .padding(paddingValues)
                .padding(horizontal = 16.dp),
            state = listState
        ) {
            items(messages, key = { it.id }) { message ->
                MessageBubble(message = message)
                Spacer(modifier = Modifier.height(8.dp))
            }
        }
    }

    // Room info dialog
    if (showInfo) {
        RoomInfoDialog(
            room = ephemeralRoom,
            onDismiss = { showInfo = false }
        )
    }

    // Extend lifetime dialog
    if (showExtendLifetime) {
        ExtendLifetimeDialog(
            room = ephemeralRoom,
            chatService = chatService,
            onDismiss = { showExtendLifetime = false }
        )
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ChatTopBar(
    room: MatrixEphemeralRoom,
    onBack: () -> Unit,
    onInfo: () -> Unit,
    onExtend: () -> Unit
) {
    Column {
        TopAppBar(
            title = { Text("Anonymous Chat") },
            navigationIcon = {
                IconButton(onClick = onBack) {
                    Icon(Icons.Default.ArrowBack, "Back")
                }
            },
            actions = {
                IconButton(onClick = onInfo) {
                    Icon(Icons.Default.Info, "Info")
                }
            }
        )

        // Expiry banner
        RoomExpiryBanner(
            room = room,
            onExtend = onExtend
        )

        Divider()
    }
}

@Composable
fun RoomExpiryBanner(
    room: MatrixEphemeralRoom,
    onExtend: () -> Unit
) {
    val isExpiringSoon = room.isExpiringSoon
    val backgroundColor = if (isExpiringSoon) {
        MaterialTheme.colorScheme.errorContainer
    } else {
        MaterialTheme.colorScheme.surfaceVariant
    }

    Row(
        modifier = Modifier
            .fillMaxWidth()
            .background(backgroundColor)
            .padding(12.dp),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically
    ) {
        Row(verticalAlignment = Alignment.CenterVertically) {
            Icon(
                imageVector = if (isExpiringSoon) {
                    androidx.compose.material.icons.Icons.Default.Warning
                } else {
                    androidx.compose.material.icons.Icons.Default.Schedule
                },
                contentDescription = null,
                tint = if (isExpiringSoon) {
                    MaterialTheme.colorScheme.error
                } else {
                    MaterialTheme.colorScheme.primary
                }
            )

            Spacer(modifier = Modifier.width(8.dp))

            Column {
                Text(
                    text = "Ephemeral Chat",
                    style = MaterialTheme.typography.labelSmall
                )
                Text(
                    text = "Expires in ${room.timeRemainingFormatted}",
                    style = MaterialTheme.typography.bodyMedium,
                    fontWeight = FontWeight.Medium,
                    color = if (isExpiringSoon) {
                        MaterialTheme.colorScheme.error
                    } else {
                        MaterialTheme.colorScheme.onSurface
                    }
                )
            }
        }

        if (isExpiringSoon) {
            Button(
                onClick = onExtend,
                modifier = Modifier.height(32.dp),
                contentPadding = PaddingValues(horizontal = 12.dp)
            ) {
                Text("Extend", fontSize = 12.sp)
            }
        }
    }
}

@Composable
fun MessageBubble(message: MessageItem) {
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = if (message.isMine) Arrangement.End else Arrangement.Start
    ) {
        if (!message.isMine) {
            Spacer(modifier = Modifier.width(48.dp))
        }

        Column(
            horizontalAlignment = if (message.isMine) Alignment.End else Alignment.Start
        ) {
            // Message bubble
            Surface(
                shape = RoundedCornerShape(16.dp),
                color = if (message.isMine) {
                    MaterialTheme.colorScheme.primary
                } else {
                    MaterialTheme.colorScheme.surfaceVariant
                }
            ) {
                Text(
                    text = message.text,
                    modifier = Modifier.padding(horizontal = 12.dp, vertical = 8.dp),
                    color = if (message.isMine) {
                        MaterialTheme.colorScheme.onPrimary
                    } else {
                        MaterialTheme.colorScheme.onSurface
                    }
                )
            }

            // Timestamp and status
            Row(
                modifier = Modifier.padding(horizontal = 4.dp, vertical = 2.dp),
                horizontalArrangement = Arrangement.spacedBy(4.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = message.timeFormatted,
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )

                if (message.isMine) {
                    if (message.isSent) {
                        Icon(
                            imageVector = androidx.compose.material.icons.Icons.Default.Check,
                            contentDescription = "Sent",
                            modifier = Modifier.size(12.dp),
                            tint = MaterialTheme.colorScheme.primary
                        )
                    } else {
                        CircularProgressIndicator(
                            modifier = Modifier.size(12.dp),
                            strokeWidth = 2.dp
                        )
                    }
                }
            }
        }

        if (message.isMine) {
            Spacer(modifier = Modifier.width(48.dp))
        }
    }
}

@Composable
fun MessageInputBar(
    text: String,
    onTextChange: (String) -> Unit,
    onSend: () -> Unit
) {
    Surface(
        modifier = Modifier.fillMaxWidth(),
        color = MaterialTheme.colorScheme.surface,
        tonalElevation = 3.dp
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(8.dp),
            verticalAlignment = Alignment.Bottom
        ) {
            OutlinedTextField(
                value = text,
                onValueChange = onTextChange,
                modifier = Modifier
                    .weight(1f)
                    .heightIn(min = 48.dp, max = 120.dp),
                placeholder = { Text("Message") },
                shape = RoundedCornerShape(24.dp),
                maxLines = 5
            )

            Spacer(modifier = Modifier.width(8.dp))

            IconButton(
                onClick = onSend,
                enabled = text.isNotBlank()
            ) {
                Icon(
                    imageVector = Icons.Default.Send,
                    contentDescription = "Send",
                    tint = if (text.isNotBlank()) {
                        MaterialTheme.colorScheme.primary
                    } else {
                        MaterialTheme.colorScheme.onSurface.copy(alpha = 0.38f)
                    }
                )
            }
        }
    }
}

@Composable
fun RoomInfoDialog(
    room: MatrixEphemeralRoom,
    onDismiss: () -> Unit
) {
    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("Chat Info") },
        text = {
            Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
                InfoRow("Room ID", room.roomId)
                InfoRow("Messages", room.messageCount.toString())
                InfoRow("Created", formatDate(room.createdAt))
                InfoRow("Expires At", formatDate(room.expiresAt))
                InfoRow("Time Remaining", room.timeRemainingFormatted)

                Divider(modifier = Modifier.padding(vertical = 8.dp))

                Text(
                    text = "Privacy & Security",
                    style = MaterialTheme.typography.titleSmall,
                    fontWeight = FontWeight.Bold
                )

                PrivacyFeature("🔒", "End-to-End Encrypted", "Messages encrypted with MLS")
                PrivacyFeature("⏱️", "Auto-Delete", "Messages deleted when trip ends")
                PrivacyFeature("👤", "Anonymous", "No personal information shared")
            }
        },
        confirmButton = {
            TextButton(onClick = onDismiss) {
                Text("Close")
            }
        }
    )
}

@Composable
fun InfoRow(label: String, value: String) {
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.SpaceBetween
    ) {
        Text(
            text = label,
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
        Text(
            text = value,
            style = MaterialTheme.typography.bodyMedium
        )
    }
}

@Composable
fun PrivacyFeature(icon: String, title: String, description: String) {
    Row(
        horizontalArrangement = Arrangement.spacedBy(8.dp),
        verticalAlignment = Alignment.Top
    ) {
        Text(icon, fontSize = 16.sp)
        Column {
            Text(
                text = title,
                style = MaterialTheme.typography.bodyMedium,
                fontWeight = FontWeight.Medium
            )
            Text(
                text = description,
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }
    }
}

@Composable
fun ExtendLifetimeDialog(
    room: MatrixEphemeralRoom,
    chatService: MatrixChatService,
    onDismiss: () -> Unit
) {
    var selectedHours by remember { mutableStateOf(1) }
    var isExtending by remember { mutableStateOf(false) }
    val scope = rememberCoroutineScope()

    val hourOptions = listOf(1, 2, 4, 8, 12, 24)

    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("Extend Chat Lifetime") },
        text = {
            Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
                // Current expiry
                Card(
                    modifier = Modifier.fillMaxWidth(),
                    colors = CardDefaults.cardColors(
                        containerColor = MaterialTheme.colorScheme.surfaceVariant
                    )
                ) {
                    Column(
                        modifier = Modifier.padding(16.dp),
                        horizontalAlignment = Alignment.CenterHorizontally
                    ) {
                        Text(
                            text = "Current Expiry",
                            style = MaterialTheme.typography.labelMedium
                        )
                        Text(
                            text = formatDate(room.expiresAt),
                            style = MaterialTheme.typography.titleMedium,
                            fontWeight = FontWeight.Bold
                        )
                        Text(
                            text = "(${room.timeRemainingFormatted} remaining)",
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.error
                        )
                    }
                }

                // Hour selector
                Text("Extend by:", style = MaterialTheme.typography.titleSmall)

                hourOptions.chunked(3).forEach { rowOptions ->
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.spacedBy(8.dp)
                    ) {
                        rowOptions.forEach { hours ->
                            FilterChip(
                                selected = selectedHours == hours,
                                onClick = { selectedHours = hours },
                                label = { Text("${hours}h") },
                                modifier = Modifier.weight(1f)
                            )
                        }
                    }
                }

                // New expiry preview
                Card(
                    modifier = Modifier.fillMaxWidth(),
                    colors = CardDefaults.cardColors(
                        containerColor = MaterialTheme.colorScheme.primaryContainer
                    )
                ) {
                    Column(
                        modifier = Modifier.padding(16.dp),
                        horizontalAlignment = Alignment.CenterHorizontally
                    ) {
                        Text(
                            text = "New Expiry",
                            style = MaterialTheme.typography.labelMedium
                        )
                        Text(
                            text = formatDate(
                                Date(room.expiresAt.time + selectedHours * 3600 * 1000L)
                            ),
                            style = MaterialTheme.typography.titleMedium,
                            fontWeight = FontWeight.Bold,
                            color = MaterialTheme.colorScheme.primary
                        )
                    }
                }
            }
        },
        confirmButton = {
            Button(
                onClick = {
                    scope.launch {
                        isExtending = true
                        try {
                            chatService.extendRoomLifetime(room, selectedHours)
                            onDismiss()
                        } catch (e: Exception) {
                            // Handle error
                        } finally {
                            isExtending = false
                        }
                    }
                },
                enabled = !isExtending
            ) {
                if (isExtending) {
                    CircularProgressIndicator(
                        modifier = Modifier.size(16.dp),
                        strokeWidth = 2.dp
                    )
                } else {
                    Text("Extend")
                }
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) {
                Text("Cancel")
            }
        }
    )
}

private fun formatDate(date: Date): String {
    val formatter = SimpleDateFormat("MMM dd, HH:mm", Locale.getDefault())
    return formatter.format(date)
}
