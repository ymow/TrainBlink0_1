package org.trainblink.app

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowForward
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import androidx.navigation.NavType
import androidx.navigation.navArgument
import org.trainblink.app.network.MatrixEphemeralRoom
import org.trainblink.app.services.MatrixChatService
import org.trainblink.app.ui.ChatListScreen
import org.trainblink.app.ui.ChatMessageScreen
import java.util.UUID

class MainActivity : ComponentActivity() {
    
    private lateinit var chatService: MatrixChatService
    
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        
        // Initialize Matrix chat service
        chatService = MatrixChatService(this, UUID.randomUUID())
        
        setContent {
            TrainBlinkTheme {
                Surface(
                    modifier = Modifier.fillMaxSize(),
                    color = MaterialTheme.colorScheme.background
                ) {
                    TrainBlinkApp(chatService = chatService)
                }
            }
        }
    }
}

@Composable
fun TrainBlinkApp(chatService: MatrixChatService) {
    val navController = rememberNavController()
    var showWelcome by remember { mutableStateOf(true) }
    
    if (showWelcome) {
        WelcomeScreen(
            onGetStarted = { 
                showWelcome = false 
            }
        )
    } else {
        NavHost(
            navController = navController,
            startDestination = "chat_list"
        ) {
            composable("chat_list") {
                ChatListScreen(
                    chatService = chatService,
                    onChatClick = { room ->
                        navController.navigate("chat_message/${room.roomId}")
                    }
                )
            }
            
            composable(
                "chat_message/{roomId}",
                arguments = listOf(navArgument("roomId") { type = NavType.StringType })
            ) { backStackEntry ->
                val roomId = backStackEntry.arguments?.getString("roomId") ?: ""
                ChatMessageScreen(
                    chatService = chatService,
                    roomId = roomId,
                    onBackClick = {
                        navController.popBackStack()
                    }
                )
            }
        }
    }
}

@Composable
fun WelcomeScreen(onGetStarted: () -> Unit = {}) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(32.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        Text(
            text = "🚂",
            fontSize = 72.sp,
            textAlign = TextAlign.Center
        )
        
        Spacer(modifier = Modifier.height(24.dp))
        
        Text(
            text = "TrainBlink",
            style = MaterialTheme.typography.headlineLarge,
            fontWeight = FontWeight.Bold,
            textAlign = TextAlign.Center
        )
        
        Spacer(modifier = Modifier.height(16.dp))
        
        Text(
            text = "Connect with fellow travelers",
            style = MaterialTheme.typography.bodyLarge,
            textAlign = TextAlign.Center,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
        
        Spacer(modifier = Modifier.height(32.dp))
        
        Button(
            onClick = onGetStarted,
            modifier = Modifier.fillMaxWidth()
        ) {
            Text("Get Started")
            Spacer(modifier = Modifier.width(8.dp))
            Icon(
                imageVector = Icons.Default.ArrowForward,
                contentDescription = "Start"
            )
        }
        
        Spacer(modifier = Modifier.height(32.dp))
        
        Card(
            modifier = Modifier.fillMaxWidth(),
            colors = CardDefaults.cardColors(
                containerColor = MaterialTheme.colorScheme.secondaryContainer
            )
        ) {
            Column(
                modifier = Modifier.padding(16.dp),
                horizontalAlignment = Alignment.CenterHorizontally
            ) {
                Text(
                    text = "🔗 Matrix Integration",
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.Bold
                )
                
                Spacer(modifier = Modifier.height(8.dp))
                
                Text(
                    text = "• Decentralized messaging\n• End-to-end encryption\n• Ephemeral chat rooms",
                    style = MaterialTheme.typography.bodyMedium,
                    textAlign = TextAlign.Start
                )
            }
        }
    }
}


@Composable
fun TrainBlinkTheme(
    content: @Composable () -> Unit
) {
    MaterialTheme(
        content = content
    )
}

@Preview(showBackground = true)
@Composable
fun WelcomeScreenPreview() {
    TrainBlinkTheme {
        WelcomeScreen()
    }
}