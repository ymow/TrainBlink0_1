# Geofence Integration Flow - Complete Architecture

**Date**: 2025-11-20
**Purpose**: Define how geofencing interacts with mobile apps, backend API, and Matrix

---

## 📱 Complete User Journey

### Scenario: User Approaches Tokyo Station

```
1. User walking → Tokyo Station (200m away)
2. GPS detects entry into geofence boundary (100m radius)
3. Mobile app triggers "station_enter" event
4. App sends event to backend API
5. Backend creates/joins Matrix room for Tokyo Station
6. Backend returns Matrix credentials to app
7. App initializes Matrix SDK with credentials
8. App connects to Tokyo Station room
9. User starts receiving content (icebreaker cards, messages)
10. User leaves station
11. GPS detects exit from geofence
12. App sends "station_exit" event
13. Backend removes user from Matrix room
14. App disconnects from Matrix room
```

---

## 🏗 System Architecture

### Three-Layer Architecture

```
┌────────────────────────────────────────────────────────────┐
│                    MOBILE APP LAYER                         │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ Geofence Manager (iOS/Android)                       │  │
│  │                                                       │  │
│  │ - Core Location (iOS) / Geofencing API (Android)    │  │
│  │ - Monitor station geofences                          │  │
│  │ - Trigger events on enter/exit                       │  │
│  │ - Background location updates                        │  │
│  └────────────┬─────────────────────────────────────────┘  │
│               │                                             │
│               │ When enter/exit detected:                   │
│               │ POST /api/v1/geofence/enter|exit           │
│               │                                             │
│  ┌────────────▼─────────────────────────────────────────┐  │
│  │ Matrix SDK (iOS/Android)                             │  │
│  │                                                       │  │
│  │ - Initialize with credentials from backend           │  │
│  │ - Connect to station room                            │  │
│  │ - Receive real-time messages                         │  │
│  │ - Send messages                                       │  │
│  │ - Handle E2EE (MLS)                                   │  │
│  └──────────────────────────────────────────────────────┘  │
└───────────────────────┬────────────────────────────────────┘
                        │
                        │ HTTPS/REST
                        ↓
┌────────────────────────────────────────────────────────────┐
│                   BACKEND API LAYER                         │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ Geofence API Handler                                 │  │
│  │ POST /api/v1/geofence/enter                          │  │
│  │ POST /api/v1/geofence/exit                           │  │
│  └────────────┬─────────────────────────────────────────┘  │
│               │                                             │
│               ↓                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ Geofence Service                                     │  │
│  │                                                       │  │
│  │ - Validate station exists                            │  │
│  │ - Validate user location                             │  │
│  │ - Check geofence boundaries                          │  │
│  │ - Rate limiting                                       │  │
│  │ - Analytics tracking                                 │  │
│  └────────────┬─────────────────────────────────────────┘  │
│               │                                             │
│               ↓                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ Matrix Bridge Service                                │  │
│  │                                                       │  │
│  │ • User Manager:                                      │  │
│  │   - Get or create Matrix user                        │  │
│  │   - Generate access token                            │  │
│  │   - Store user mapping (Firebase UID → MXID)        │  │
│  │                                                       │  │
│  │ • Room Manager:                                      │  │
│  │   - Get or create station room                       │  │
│  │   - Add user to room                                 │  │
│  │   - Remove user from room                            │  │
│  │   - Update member count                              │  │
│  │                                                       │  │
│  │ • Content Publisher:                                 │  │
│  │   - Publish icebreaker cards to room                │  │
│  │   - Send announcements                               │  │
│  └────────────┬─────────────────────────────────────────┘  │
│               │                                             │
│               ↓                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ Response to Mobile App                               │  │
│  │                                                       │  │
│  │ {                                                    │  │
│  │   "matrix": {                                        │  │
│  │     "user_id": "@trainblink_abc:trainblink.org",    │  │
│  │     "access_token": "syt_...",                       │  │
│  │     "device_id": "DEVICE123",                        │  │
│  │     "homeserver_url": "https://matrix.trainblink.org",│ │
│  │     "room_id": "!tokyo001:trainblink.org",          │  │
│  │     "room_alias": "#tokyo-station:trainblink.org"   │  │
│  │   }                                                  │  │
│  │ }                                                    │  │
│  └──────────────────────────────────────────────────────┘  │
└───────────────────────┬────────────────────────────────────┘
                        │
                        │ Matrix Client-Server API
                        ↓
┌────────────────────────────────────────────────────────────┐
│                   MATRIX LAYER                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ Matrix Synapse Homeserver                            │  │
│  │                                                       │  │
│  │ Room: !tokyo001:trainblink.org                       │  │
│  │ Alias: #tokyo-station:trainblink.org                 │  │
│  │                                                       │  │
│  │ Members:                                             │  │
│  │ - @trainblink_user1:trainblink.org                   │  │
│  │ - @trainblink_user2:trainblink.org                   │  │
│  │ - @trainblink_user3:trainblink.org                   │  │
│  │                                                       │  │
│  │ Events:                                              │  │
│  │ - m.room.message (icebreaker cards)                  │  │
│  │ - m.room.member (join/leave)                         │  │
│  │ - m.typing (typing indicators)                       │  │
│  └──────────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────────┘
```

---

## 🔄 Detailed Interaction Flows

### Flow 1: Station Entry (Geofence Enter)

```sequence
Mobile App → Backend: POST /api/v1/geofence/enter
                      {
                        station_id: "station_tokyo_001",
                        latitude: 35.6812,
                        longitude: 139.7671,
                        timestamp: "2025-11-20T10:00:00Z"
                      }

Backend → Geofence Service: Validate station & location
Geofence Service → Database: Check station exists
Database → Geofence Service: Station data
Geofence Service → Geofence Service: Verify GPS within 100m radius

Geofence Service → Matrix Bridge: HandleStationEnter(user_id, station_id)

Matrix Bridge → User Manager: GetOrCreateMatrixUser(user_id)
User Manager → Database: SELECT from matrix_users WHERE user_id = ?
Database → User Manager: Not found
User Manager → Matrix Synapse: POST /_matrix/client/v3/register
                                {
                                  username: "trainblink_abc123",
                                  password: "generated_password",
                                  auth: { type: "m.login.application_service" }
                                }
Matrix Synapse → User Manager: {
                                 user_id: "@trainblink_abc123:trainblink.org",
                                 access_token: "syt_xxx",
                                 device_id: "DEVICE123"
                               }
User Manager → Database: INSERT into matrix_users
Database → User Manager: OK
User Manager → Matrix Bridge: MatrixUser object

Matrix Bridge → Room Manager: GetOrCreateStationRoom(station_id)
Room Manager → Database: SELECT from matrix_rooms WHERE station_id = ?
Database → Room Manager: Room found (!tokyo001:trainblink.org)
Room Manager → Matrix Bridge: Room object

Matrix Bridge → Room Manager: JoinUserToRoom(user_id, room_id)
Room Manager → Matrix Synapse: POST /_matrix/client/v3/rooms/{room_id}/join
                                Authorization: Bearer {bot_token}
                                {
                                  user_id: "@trainblink_abc123:trainblink.org"
                                }
Matrix Synapse → Room Manager: { room_id: "!tokyo001:trainblink.org" }
Room Manager → Database: INSERT into matrix_room_memberships
                         {
                           room_id: (uuid),
                           user_id: (uuid),
                           membership: "join",
                           joined_at: NOW()
                         }
Room Manager → Database: UPDATE matrix_rooms SET member_count = member_count + 1
Database → Room Manager: OK
Room Manager → Matrix Bridge: OK

Matrix Bridge → Backend: {
                           matrix_user_id: "@trainblink_abc123:trainblink.org",
                           access_token: "syt_xxx",
                           device_id: "DEVICE123",
                           room_id: "!tokyo001:trainblink.org",
                           room_alias: "#tokyo-station:trainblink.org"
                         }

Backend → Mobile App: HTTP 200 OK
                      {
                        status: "success",
                        data: {
                          station: {
                            id: "station_tokyo_001",
                            name: "Tokyo Station"
                          },
                          matrix: {
                            user_id: "@trainblink_abc123:trainblink.org",
                            access_token: "syt_xxx",
                            device_id: "DEVICE123",
                            homeserver_url: "https://matrix.trainblink.org",
                            room_id: "!tokyo001:trainblink.org",
                            room_alias: "#tokyo-station:trainblink.org"
                          }
                        }
                      }

Mobile App → Matrix SDK: Initialize
                         MatrixClient(
                           homeserver: "https://matrix.trainblink.org",
                           userId: "@trainblink_abc123:trainblink.org",
                           accessToken: "syt_xxx",
                           deviceId: "DEVICE123"
                         )

Mobile App → Matrix SDK: client.joinRoom("!tokyo001:trainblink.org")
Matrix SDK → Matrix Synapse: POST /_matrix/client/v3/rooms/!tokyo001/join
Matrix Synapse → Matrix SDK: { room_id: "!tokyo001:trainblink.org" }

Mobile App → Matrix SDK: client.startSync()
Matrix SDK → Matrix Synapse: GET /_matrix/client/v3/sync?timeout=30000
Matrix Synapse → Matrix SDK: {
                               rooms: {
                                 join: {
                                   "!tokyo001:trainblink.org": {
                                     timeline: { events: [...] },
                                     state: { events: [...] }
                                   }
                                 }
                               }
                             }

Mobile App → User: Display "Joined Tokyo Station"
Mobile App → User: Show icebreaker cards feed
Mobile App → User: Enable real-time chat
```

---

### Flow 2: Station Exit (Geofence Exit)

```sequence
Mobile App → Backend: POST /api/v1/geofence/exit
                      {
                        station_id: "station_tokyo_001",
                        latitude: 35.6900,
                        longitude: 139.7800,
                        timestamp: "2025-11-20T11:00:00Z"
                      }

Backend → Geofence Service: Validate exit event
Geofence Service → Database: Check user was in station
Database → Geofence Service: User was in station (confirmed)

Geofence Service → Matrix Bridge: HandleStationExit(user_id, station_id)

Matrix Bridge → Room Manager: RemoveUserFromRoom(user_id, station_id)
Room Manager → Database: SELECT matrix_user_id from matrix_users WHERE user_id = ?
Database → Room Manager: @trainblink_abc123:trainblink.org
Room Manager → Database: SELECT room_id from matrix_rooms WHERE station_id = ?
Database → Room Manager: !tokyo001:trainblink.org

Room Manager → Matrix Synapse: POST /_matrix/client/v3/rooms/{room_id}/leave
                                Authorization: Bearer {user_access_token}
Matrix Synapse → Room Manager: {}
Room Manager → Database: UPDATE matrix_room_memberships
                         SET membership = 'leave', left_at = NOW()
                         WHERE room_id = ? AND user_id = ?
Room Manager → Database: UPDATE matrix_rooms SET member_count = member_count - 1
Database → Room Manager: OK
Room Manager → Matrix Bridge: OK

Matrix Bridge → Backend: { left: true }

Backend → Mobile App: HTTP 200 OK
                      {
                        status: "success",
                        data: {
                          station: {
                            id: "station_tokyo_001",
                            name: "Tokyo Station"
                          },
                          matrix: {
                            left: true
                          }
                        }
                      }

Mobile App → Matrix SDK: client.leaveRoom("!tokyo001:trainblink.org")
Matrix SDK → Matrix Synapse: POST /_matrix/client/v3/rooms/!tokyo001/leave
Matrix Synapse → Matrix SDK: {}

Mobile App → Matrix SDK: client.stopSync()
Mobile App → User: Display "Left Tokyo Station"
Mobile App → User: Hide station feed
```

---

## 📱 Mobile App Implementation

### iOS Example (Swift + CoreLocation)

```swift
import CoreLocation
import MatrixSDK

class GeofenceManager: NSObject, CLLocationManagerDelegate {
    private let locationManager = CLLocationManager()
    private let apiClient: APIClient
    private var matrixClient: MXRestClient?
    private var currentStation: Station?

    // Station geofences (loaded from backend)
    private var monitoredStations: [Station] = []

    override init() {
        self.apiClient = APIClient()
        super.init()

        locationManager.delegate = self
        locationManager.desiredAccuracy = kCLLocationAccuracyBest
        locationManager.allowsBackgroundLocationUpdates = true

        // Request permissions
        locationManager.requestAlwaysAuthorization()
    }

    // Load stations and set up geofences
    func setupGeofences() async {
        // Get stations from backend
        let stations = try? await apiClient.getStations()

        stations?.forEach { station in
            let region = CLCircularRegion(
                center: CLLocationCoordinate2D(
                    latitude: station.latitude,
                    longitude: station.longitude
                ),
                radius: 100.0, // 100 meters
                identifier: station.id
            )

            region.notifyOnEntry = true
            region.notifyOnExit = true

            locationManager.startMonitoring(for: region)
            monitoredStations.append(station)
        }
    }

    // MARK: - CLLocationManagerDelegate

    // Called when entering a geofence
    func locationManager(_ manager: CLLocationManager, didEnterRegion region: CLRegion) {
        guard let station = monitoredStations.first(where: { $0.id == region.identifier }) else {
            return
        }

        Task {
            await handleStationEnter(station: station)
        }
    }

    // Called when exiting a geofence
    func locationManager(_ manager: CLLocationManager, didExitRegion region: CLRegion) {
        guard let station = monitoredStations.first(where: { $0.id == region.identifier }) else {
            return
        }

        Task {
            await handleStationExit(station: station)
        }
    }

    // MARK: - Station Entry

    private func handleStationEnter(station: Station) async {
        print("📍 Entered station: \(station.name)")

        currentStation = station

        do {
            // Call backend API
            let response = try await apiClient.enterStation(
                stationId: station.id,
                latitude: station.latitude,
                longitude: station.longitude
            )

            // Initialize Matrix SDK with credentials
            await initializeMatrixClient(
                homeserverURL: response.matrix.homeserverUrl,
                userId: response.matrix.userId,
                accessToken: response.matrix.accessToken,
                deviceId: response.matrix.deviceId
            )

            // Join station room
            await joinStationRoom(roomId: response.matrix.roomId)

            // Notify UI
            NotificationCenter.default.post(
                name: .didEnterStation,
                object: nil,
                userInfo: ["station": station, "matrix": response.matrix]
            )

            // Show local notification
            showNotification(
                title: "Welcome to \(station.name)!",
                body: "You've joined the station chat. Check out icebreaker cards!"
            )

        } catch {
            print("❌ Failed to enter station: \(error)")
        }
    }

    // MARK: - Station Exit

    private func handleStationExit(station: Station) async {
        print("📍 Left station: \(station.name)")

        do {
            // Call backend API
            let response = try await apiClient.exitStation(
                stationId: station.id
            )

            // Leave Matrix room
            if let roomId = currentStation?.matrixRoomId {
                await leaveStationRoom(roomId: roomId)
            }

            // Stop Matrix sync
            matrixClient = nil
            currentStation = nil

            // Notify UI
            NotificationCenter.default.post(
                name: .didExitStation,
                object: nil,
                userInfo: ["station": station]
            )

        } catch {
            print("❌ Failed to exit station: \(error)")
        }
    }

    // MARK: - Matrix Integration

    private func initializeMatrixClient(
        homeserverURL: String,
        userId: String,
        accessToken: String,
        deviceId: String
    ) async {
        let credentials = MXCredentials(
            homeServer: homeserverURL,
            userId: userId,
            accessToken: accessToken
        )
        credentials.deviceId = deviceId

        matrixClient = MXRestClient(credentials: credentials)

        // Start syncing
        matrixClient?.startSync()
    }

    private func joinStationRoom(roomId: String) async {
        try? await matrixClient?.joinRoom(roomId)
        print("✅ Joined Matrix room: \(roomId)")
    }

    private func leaveStationRoom(roomId: String) async {
        try? await matrixClient?.leaveRoom(roomId)
        print("✅ Left Matrix room: \(roomId)")
    }

    private func showNotification(title: String, body: String) {
        let content = UNMutableNotificationContent()
        content.title = title
        content.body = body
        content.sound = .default

        let request = UNNotificationRequest(
            identifier: UUID().uuidString,
            content: content,
            trigger: nil
        )

        UNUserNotificationCenter.current().add(request)
    }
}

// MARK: - API Models

struct EnterStationResponse: Codable {
    let status: String
    let data: StationData

    struct StationData: Codable {
        let station: Station
        let matrix: MatrixCredentials
    }
}

struct MatrixCredentials: Codable {
    let userId: String
    let accessToken: String
    let deviceId: String
    let homeserverUrl: String
    let roomId: String
    let roomAlias: String

    enum CodingKeys: String, CodingKey {
        case userId = "user_id"
        case accessToken = "access_token"
        case deviceId = "device_id"
        case homeserverUrl = "homeserver_url"
        case roomId = "room_id"
        case roomAlias = "room_alias"
    }
}
```

---

### Android Example (Kotlin + Geofencing API)

```kotlin
import android.app.PendingIntent
import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import com.google.android.gms.location.Geofence
import com.google.android.gms.location.GeofencingClient
import com.google.android.gms.location.GeofencingRequest
import com.google.android.gms.location.LocationServices
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch

class GeofenceManager(private val context: Context) {
    private val geofencingClient: GeofencingClient = LocationServices.getGeofencingClient(context)
    private val apiClient = APIClient()
    private var matrixClient: MatrixClient? = null

    // Set up geofences for all stations
    suspend fun setupGeofences() {
        val stations = apiClient.getStations()

        val geofences = stations.map { station ->
            Geofence.Builder()
                .setRequestId(station.id)
                .setCircularRegion(station.latitude, station.longitude, 100f) // 100m radius
                .setExpirationDuration(Geofence.NEVER_EXPIRE)
                .setTransitionTypes(Geofence.GEOFENCE_TRANSITION_ENTER or Geofence.GEOFENCE_TRANSITION_EXIT)
                .build()
        }

        val geofencingRequest = GeofencingRequest.Builder()
            .setInitialTrigger(GeofencingRequest.INITIAL_TRIGGER_ENTER)
            .addGeofences(geofences)
            .build()

        val pendingIntent = getGeofencePendingIntent()

        geofencingClient.addGeofences(geofencingRequest, pendingIntent)
    }

    private fun getGeofencePendingIntent(): PendingIntent {
        val intent = Intent(context, GeofenceBroadcastReceiver::class.java)
        return PendingIntent.getBroadcast(
            context,
            0,
            intent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_MUTABLE
        )
    }
}

// Broadcast receiver for geofence events
class GeofenceBroadcastReceiver : BroadcastReceiver() {
    override fun onReceive(context: Context, intent: Intent) {
        val geofencingEvent = GeofencingEvent.fromIntent(intent) ?: return

        if (geofencingEvent.hasError()) {
            Log.e("Geofence", "Error: ${geofencingEvent.errorCode}")
            return
        }

        val transition = geofencingEvent.geofenceTransition
        val triggeringGeofences = geofencingEvent.triggeringGeofences ?: return

        CoroutineScope(Dispatchers.IO).launch {
            triggeringGeofences.forEach { geofence ->
                val stationId = geofence.requestId

                when (transition) {
                    Geofence.GEOFENCE_TRANSITION_ENTER -> {
                        handleStationEnter(context, stationId)
                    }
                    Geofence.GEOFENCE_TRANSITION_EXIT -> {
                        handleStationExit(context, stationId)
                    }
                }
            }
        }
    }

    private suspend fun handleStationEnter(context: Context, stationId: String) {
        Log.d("Geofence", "Entered station: $stationId")

        try {
            val apiClient = APIClient()
            val response = apiClient.enterStation(
                stationId = stationId,
                latitude = 0.0, // Get from location
                longitude = 0.0
            )

            // Initialize Matrix client
            val matrixClient = MatrixClient(
                homeserverUrl = response.data.matrix.homeserverUrl,
                userId = response.data.matrix.userId,
                accessToken = response.data.matrix.accessToken,
                deviceId = response.data.matrix.deviceId
            )

            // Join room
            matrixClient.joinRoom(response.data.matrix.roomId)

            // Show notification
            showNotification(
                context,
                "Welcome to ${response.data.station.name}!",
                "You've joined the station chat"
            )

        } catch (e: Exception) {
            Log.e("Geofence", "Failed to enter station", e)
        }
    }

    private suspend fun handleStationExit(context: Context, stationId: String) {
        Log.d("Geofence", "Left station: $stationId")

        try {
            val apiClient = APIClient()
            apiClient.exitStation(stationId)

            // Leave Matrix room and cleanup
            // ...

        } catch (e: Exception) {
            Log.e("Geofence", "Failed to exit station", e)
        }
    }
}
```

---

## 🗄 Backend Implementation

### Geofence API Handler (Go)

```go
// internal/api/geofence_handler.go

package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ymow/messenger_protocol_research/internal/geofence"
	"github.com/ymow/messenger_protocol_research/internal/matrix"
)

type GeofenceHandler struct {
	geofenceService *geofence.Service
	matrixBridge    *matrix.GeofenceBridge
}

func NewGeofenceHandler(
	geofenceService *geofence.Service,
	matrixBridge *matrix.GeofenceBridge,
) *GeofenceHandler {
	return &GeofenceHandler{
		geofenceService: geofenceService,
		matrixBridge:    matrixBridge,
	}
}

// POST /api/v1/geofence/enter
func (h *GeofenceHandler) EnterStation(c *gin.Context) {
	// Get authenticated user
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		StationID string  `json:"station_id" binding:"required"`
		Latitude  float64 `json:"latitude" binding:"required"`
		Longitude float64 `json:"longitude" binding:"required"`
		Timestamp string  `json:"timestamp"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate station entry
	station, err := h.geofenceService.ValidateStationEntry(
		c.Request.Context(),
		userID.(string),
		req.StationID,
		req.Latitude,
		req.Longitude,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Trigger Matrix integration
	matrixCredentials, err := h.matrixBridge.HandleStationEnter(
		c.Request.Context(),
		userID.(string),
		req.StationID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to join station room",
		})
		return
	}

	// Record entry in database
	if err := h.geofenceService.RecordEntry(
		c.Request.Context(),
		userID.(string),
		req.StationID,
		req.Latitude,
		req.Longitude,
	); err != nil {
		// Log but don't fail
		log.Printf("Failed to record entry: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"station": gin.H{
				"id":   station.ID,
				"name": station.Name,
			},
			"matrix": matrixCredentials,
		},
	})
}

// POST /api/v1/geofence/exit
func (h *GeofenceHandler) ExitStation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		StationID string  `json:"station_id" binding:"required"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Timestamp string  `json:"timestamp"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Trigger Matrix integration
	if err := h.matrixBridge.HandleStationExit(
		c.Request.Context(),
		userID.(string),
		req.StationID,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to leave station room",
		})
		return
	}

	// Record exit in database
	if err := h.geofenceService.RecordExit(
		c.Request.Context(),
		userID.(string),
		req.StationID,
	); err != nil {
		log.Printf("Failed to record exit: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"station_id": req.StationID,
			"left":       true,
		},
	})
}
```

### Matrix Geofence Bridge (Go)

```go
// internal/matrix/geofence_bridge.go

package matrix

import (
	"context"
	"fmt"
)

type GeofenceBridge struct {
	userManager *UserManager
	roomManager *RoomManager
}

func NewGeofenceBridge(
	userManager *UserManager,
	roomManager *RoomManager,
) *GeofenceBridge {
	return &GeofenceBridge{
		userManager: userManager,
		roomManager: roomManager,
	}
}

// HandleStationEnter handles user entering a station geofence
func (gb *GeofenceBridge) HandleStationEnter(
	ctx context.Context,
	userID string,
	stationID string,
) (*MatrixCredentials, error) {
	// Step 1: Get or create Matrix user for TrainBlink user
	matrixUser, err := gb.userManager.GetOrCreateMatrixUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get/create matrix user: %w", err)
	}

	// Step 2: Get or create station room
	room, err := gb.roomManager.GetOrCreateStationRoom(ctx, stationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get/create station room: %w", err)
	}

	// Step 3: Join user to room
	if err := gb.roomManager.JoinUserToRoom(ctx, userID, room.RoomID); err != nil {
		return nil, fmt.Errorf("failed to join user to room: %w", err)
	}

	// Step 4: Return credentials to mobile app
	return &MatrixCredentials{
		UserID:        matrixUser.MatrixUserID,
		AccessToken:   matrixUser.AccessToken,
		DeviceID:      matrixUser.DeviceID,
		HomeserverURL: "https://matrix.trainblink.org",
		RoomID:        room.RoomID,
		RoomAlias:     room.RoomAlias,
	}, nil
}

// HandleStationExit handles user leaving a station geofence
func (gb *GeofenceBridge) HandleStationExit(
	ctx context.Context,
	userID string,
	stationID string,
) error {
	// Get room for station
	room, err := gb.roomManager.GetStationRoom(ctx, stationID)
	if err != nil {
		return fmt.Errorf("failed to get station room: %w", err)
	}

	// Remove user from room
	if err := gb.roomManager.RemoveUserFromRoom(ctx, userID, room.RoomID); err != nil {
		return fmt.Errorf("failed to remove user from room: %w", err)
	}

	return nil
}

type MatrixCredentials struct {
	UserID        string `json:"user_id"`
	AccessToken   string `json:"access_token"`
	DeviceID      string `json:"device_id"`
	HomeserverURL string `json:"homeserver_url"`
	RoomID        string `json:"room_id"`
	RoomAlias     string `json:"room_alias"`
}
```

---

## 🔑 Key Design Decisions

### 1. **Geofence Radius: 100 meters**
- Balanced between accuracy and battery life
- Large enough to trigger reliably
- Small enough to be station-specific

### 2. **Backend Handles Matrix Logic**
- Mobile app doesn't need to know Matrix protocol details
- Backend returns ready-to-use credentials
- Easier to update Matrix integration without app updates

### 3. **Automatic User Creation**
- No manual Matrix registration required
- TrainBlink user → Matrix user mapping automatic
- Seamless user experience

### 4. **Background Geofencing**
- Works even when app is closed
- iOS: Core Location background updates
- Android: Geofencing API handles background

### 5. **Graceful Degradation**
- If Matrix fails, geofence still tracks entry/exit
- Analytics still work
- Can retry Matrix connection

---

## 📊 Database Tracking

### Geofence Events Table

```sql
CREATE TABLE geofence_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  user_id UUID NOT NULL REFERENCES users(id),
  station_id VARCHAR(50) NOT NULL REFERENCES stations(id),

  event_type VARCHAR(20) NOT NULL,  -- 'enter' or 'exit'

  latitude DECIMAL(10, 8),
  longitude DECIMAL(11, 8),
  accuracy DECIMAL(10, 2),          -- GPS accuracy in meters

  timestamp TIMESTAMP WITH TIME ZONE NOT NULL,

  -- Matrix integration status
  matrix_joined BOOLEAN DEFAULT false,
  matrix_room_id VARCHAR(255),

  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_geofence_events_user ON geofence_events(user_id);
CREATE INDEX idx_geofence_events_station ON geofence_events(station_id);
CREATE INDEX idx_geofence_events_timestamp ON geofence_events(timestamp);
```

---

## ✅ Summary

### Mobile App Responsibilities
- Monitor GPS location
- Detect geofence enter/exit
- Call backend API
- Initialize Matrix SDK with credentials
- Handle UI updates

### Backend Responsibilities
- Validate geofence events
- Manage Matrix users
- Manage Matrix rooms
- Handle join/leave operations
- Track analytics

### Matrix Responsibilities
- Store room state
- Route messages
- Handle E2EE (via MLS)
- Sync events to clients

---

**Does this architecture align with your understanding? Any aspects you'd like me to clarify or adjust?**
