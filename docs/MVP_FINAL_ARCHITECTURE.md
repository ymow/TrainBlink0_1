# TrainBlink MVP - Final Architecture Specification

**Version**: 3.0 (Final)
**Date**: 2025-11-20
**Status**: Ready for Implementation

---

## 🎯 MVP Vision

**Tagline**: "Meet fellow travelers. Chat ephemeral. Delete on arrival."

**User Story:**
```
As a traveler on Tokyo → Osaka Shinkansen,
I want to discover nearby travelers and chat,
So that I can have interesting conversations during my journey,
And all messages disappear when I arrive.
```

---

## ✅ Confirmed Requirements

Based on your answers:

1. ✅ **Matrix in MVP** - Cross-platform messaging (iOS ↔ Android)
2. ✅ **Ephemeral only** - Messages cached locally during trip, deleted on arrival
3. ✅ **Anonymized analytics** - Track discoveries for insights
4. ✅ **Manual trip selection** - User selects route manually
5. ✅ **BLE discovery** - 50-100m radius for user discovery
6. ✅ **MLS E2EE** - Already implemented for encryption

---

## 🏗 System Architecture

### Three-Layer Stack

```
┌─────────────────────────────────────────────────────────────┐
│                    LAYER 1: DISCOVERY                        │
│                      (Pure Local)                            │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ BLE (Bluetooth Low Energy)                             │ │
│  │                                                         │ │
│  │ Broadcasting:                                          │ │
│  │ - Anonymous ID: "TB_abc123"                            │ │
│  │ - Trip route: "Tokyo→Osaka"                            │ │
│  │ - Age range (optional): "20-25"                        │ │
│  │ - Gender (optional): "M"/"F"/"O"                       │ │
│  │                                                         │ │
│  │ Scanning:                                              │ │
│  │ - Detect users within 50-100m                          │ │
│  │ - Estimate distance via RSSI                           │ │
│  │ - Filter by route (optional)                           │ │
│  │ - Display: "5 travelers nearby on Tokyo→Osaka"         │ │
│  └────────────────────────────────────────────────────────┘ │
└────────────────────────┬────────────────────────────────────┘
                         │ User taps "Chat with TB_xyz"
                         ↓
┌─────────────────────────────────────────────────────────────┐
│                    LAYER 2: MESSAGING                        │
│                  (Matrix + MLS E2EE)                         │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Matrix Protocol                                        │ │
│  │                                                         │ │
│  │ ONLINE (Both users have internet):                     │ │
│  │   ├─ iOS ↔ Android: Matrix (only option)              │ │
│  │   └─ iOS ↔ iOS: Matrix (preferred) or AirDrop         │ │
│  │                                                         │ │
│  │ OFFLINE (No internet):                                 │ │
│  │   ├─ iOS ↔ iOS: AirDrop                               │ │
│  │   └─ Android ↔ Android: Nearby Share                  │ │
│  │                                                         │ │
│  │ Room Type: Direct Message (1-on-1)                     │ │
│  │ Encryption: MLS (RFC 9420)                             │ │
│  │ Storage: Ephemeral (auto-delete)                       │ │
│  │ Cache: Local device only during trip                   │ │
│  └────────────────────────────────────────────────────────┘ │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ↓
┌─────────────────────────────────────────────────────────────┐
│                  LAYER 3: BACKEND API                        │
│                (Minimal, Privacy-First)                      │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Go Backend Services                                    │ │
│  │                                                         │ │
│  │ 1. Trip Management Service                             │ │
│  │    - POST /api/v1/trips/start                          │ │
│  │    - POST /api/v1/trips/end                            │ │
│  │    - GET  /api/v1/trips/active                         │ │
│  │                                                         │ │
│  │ 2. Discovery Tracking Service (Analytics)              │ │
│  │    - POST /api/v1/discoveries/log (anonymized)         │ │
│  │    - GET  /api/v1/analytics/routes (admin)             │ │
│  │                                                         │ │
│  │ 3. Matrix Bridge Service                               │ │
│  │    - POST /api/v1/matrix/dm/create                     │ │
│  │    - POST /api/v1/matrix/dm/delete                     │ │
│  │    - User Manager (TrainBlink → Matrix MXID)           │ │
│  │    - Room Manager (Ephemeral DM rooms)                 │ │
│  │                                                         │ │
│  │ 4. MLS E2EE Service (Already Implemented)              │ │
│  │    - POST /api/v1/mls/groups                           │ │
│  │    - POST /api/v1/mls/keypackages                      │ │
│  │    - POST /api/v1/mls/messages                         │ │
│  └────────────────────────────────────────────────────────┘ │
└────────────────────────┬────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────────┐
│                  LAYER 4: DATA STORAGE                       │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ PostgreSQL                                             │ │
│  │ - trips (user's active trips)                          │ │
│  │ - discoveries (anonymized analytics)                   │ │
│  │ - matrix_ephemeral_rooms (auto-delete metadata)        │ │
│  │ - matrix_users (TrainBlink → MXID mapping)             │ │
│  │ - mls_groups, mls_keypackages, etc. (E2EE)            │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Redis                                                   │ │
│  │ - Active trip cache (fast lookup)                      │ │
│  │ - Rate limiting                                         │ │
│  │ - MLS sequence numbers                                  │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Matrix Synapse                                          │ │
│  │ - Ephemeral DM rooms (auto-expire in 24h)              │ │
│  │ - E2EE messages (server can't read)                    │ │
│  │ - Temporary message routing                             │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

---

## 🔄 Complete User Journey

### 1. Starting a Trip

```
┌─────────────────────────────────────────────────────────┐
│ User opens app at Tokyo Station                         │
└────────────────────┬────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────┐
│ App shows: "Select your trip"                           │
│                                                         │
│ [🚅 Tokyo → Osaka]     Nozomi 123                      │
│ [🚅 Tokyo → Kyoto]     Hikari 456                      │
│ [✈️  Tokyo → Sapporo]  NH789                           │
│                                                         │
│ Manual entry: [From: Tokyo] [To: Osaka]                │
└────────────────────┬────────────────────────────────────┘
                     ↓ User selects "Tokyo → Osaka"
┌─────────────────────────────────────────────────────────┐
│ POST /api/v1/trips/start                                │
│ {                                                       │
│   "route": "Tokyo → Osaka",                            │
│   "departure_time": "2025-11-20T10:00:00Z",            │
│   "estimated_arrival": "2025-11-20T12:30:00Z"          │
│ }                                                       │
└────────────────────┬────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────┐
│ Backend:                                                │
│ - Creates trip record                                   │
│ - Caches in Redis (trip_id → user_id)                  │
│ - Returns trip_id                                       │
└────────────────────┬────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────┐
│ App starts BLE discovery:                               │
│ - Start broadcasting (anonymous ID + route)             │
│ - Start scanning (detect nearby users)                 │
│ - Show: "Discovery enabled. 0 travelers found."        │
└─────────────────────────────────────────────────────────┘
```

### 2. Discovering Nearby Users

```
┌─────────────────────────────────────────────────────────┐
│ User A's phone (BLE)                                    │
│                                                         │
│ Broadcasting:                                           │
│ UUID: TrainBlink Service UUID                           │
│ Data: {                                                 │
│   "anonymous_id": "TB_abc123",                          │
│   "route": "Tokyo→Osaka",                               │
│   "age_range": "25-30",                                 │
│   "gender": "M"                                         │
│ }                                                       │
└────────────────────┬────────────────────────────────────┘
                     │ BLE Advertisement (50-100m radius)
                     ↓
┌─────────────────────────────────────────────────────────┐
│ User B's phone (BLE Scanning)                           │
│                                                         │
│ Detected:                                               │
│ - Anonymous ID: TB_abc123                               │
│ - Route: Tokyo→Osaka (matching!)                        │
│ - RSSI: -75 dBm → Distance: ~20m (Close)               │
│ - Age: 25-30, Gender: M                                 │
└────────────────────┬────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────┐
│ App shows:                                              │
│                                                         │
│ 🔵 Traveler_abc123                                      │
│    Tokyo → Osaka • 20m away • M, 25-30                 │
│    [💬 Chat]                                            │
│                                                         │
│ 🟢 Traveler_def456                                      │
│    Tokyo → Osaka • 50m away • F, 30-35                 │
│    [💬 Chat]                                            │
└─────────────────────────────────────────────────────────┘
```

### 3. Initiating Chat

```
User B taps "💬 Chat" on Traveler_abc123
                     ↓
┌─────────────────────────────────────────────────────────┐
│ App checks connectivity:                                │
│ - User A: Online ✅                                     │
│ - User B: Online ✅                                     │
│ → Decision: Use Matrix                                  │
└────────────────────┬────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────┐
│ POST /api/v1/matrix/dm/create                           │
│ {                                                       │
│   "with_user_anonymous_id": "TB_abc123",                │
│   "trip_id": "trip_uuid_123",                           │
│   "enable_mls": true                                    │
│ }                                                       │
└────────────────────┬────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────┐
│ Backend Matrix Bridge:                                  │
│                                                         │
│ 1. Get or create Matrix users:                         │
│    - User A: @trainblink_abc123:trainblink.org         │
│    - User B: @trainblink_def456:trainblink.org         │
│                                                         │
│ 2. Create ephemeral DM room:                            │
│    - Room ID: !dm_xyz789:trainblink.org                │
│    - Type: Direct message (is_direct: true)             │
│    - Encryption: MLS                                    │
│    - Expiry: 24 hours                                   │
│                                                         │
│ 3. Create MLS group for room:                           │
│    - Group ID: mls_group_dm_xyz                         │
│    - Members: User A, User B                            │
│    - Cipher suite: MLS_128_DHKEMX25519_AES128GCM        │
│                                                         │
│ 4. Return credentials to both users                     │
└────────────────────┬────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────┐
│ Response to User B:                                     │
│ {                                                       │
│   "room_id": "!dm_xyz789:trainblink.org",              │
│   "matrix_user_id": "@trainblink_def456:trainblink.org",│
│   "access_token": "syt_xxx",                            │
│   "device_id": "DEVICE456",                             │
│   "mls_group_id": "mls_group_dm_xyz",                   │
│   "other_user": {                                       │
│     "anonymous_id": "TB_abc123",                        │
│     "age_range": "25-30",                               │
│     "gender": "M"                                       │
│   }                                                     │
│ }                                                       │
└────────────────────┬────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────┐
│ App initializes Matrix SDK:                             │
│ - Connect to homeserver                                 │
│ - Join room                                             │
│ - Start sync                                            │
│ - Initialize MLS encryption                             │
│                                                         │
│ Chat interface opens:                                   │
│ "Chat with Traveler_abc123"                             │
│ "🔒 End-to-end encrypted"                               │
│ "⏰ Messages expire on arrival"                         │
└─────────────────────────────────────────────────────────┘
```

### 4. Messaging

```
┌─────────────────────────────────────────────────────────┐
│ User B types: "Hey! Also going to Osaka? 👋"           │
└────────────────────┬────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────┐
│ Local encryption (MLS):                                 │
│ Plaintext → MLS Encrypt → Ciphertext                    │
└────────────────────┬────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────┐
│ Send to Matrix Synapse:                                 │
│ POST /_matrix/client/v3/rooms/!dm_xyz789/send/m.room.message│
│ {                                                       │
│   "msgtype": "m.text",                                  │
│   "body": "Hey! Also going to Osaka? 👋",             │
│   "ciphertext": "base64_encrypted_data"                 │
│ }                                                       │
└────────────────────┬────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────┐
│ Matrix Synapse:                                         │
│ - Receives encrypted message (can't read it)            │
│ - Routes to User A                                      │
│ - Stores temporarily for delivery                       │
└────────────────────┬────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────┐
│ User A receives:                                        │
│ - Matrix sync delivers encrypted message                │
│ - MLS decrypts locally                                  │
│ - Shows: "Hey! Also going to Osaka? 👋"               │
└─────────────────────────────────────────────────────────┘
```

### 5. Local Caching During Trip

```
┌─────────────────────────────────────────────────────────┐
│ All messages cached locally on device:                  │
│                                                         │
│ Local SQLite Database:                                  │
│ - trip_messages table                                   │
│ - Encrypted at rest                                     │
│ - Linked to trip_id                                     │
│                                                         │
│ Why cache locally?                                      │
│ ✅ Survive temporary connection drops                   │
│ ✅ Fast message retrieval                               │
│ ✅ Resilient to network issues                          │
│                                                         │
│ When trip ends:                                         │
│ ❌ Delete local cache                                   │
│ ❌ Leave Matrix room                                    │
│ ❌ Backend auto-deletes room after 24h                  │
└─────────────────────────────────────────────────────────┘
```

### 6. Ending Trip

```
User arrives at Osaka Station
                     ↓
┌─────────────────────────────────────────────────────────┐
│ User taps "End Trip"                                    │
│ or                                                      │
│ App auto-detects (timer based on estimated arrival)     │
└────────────────────┬────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────┐
│ App shows:                                              │
│                                                         │
│ "Trip ended! 🎉"                                        │
│                                                         │
│ "You chatted with 3 travelers."                         │
│ "All messages will be deleted."                         │
│                                                         │
│ [ Keep Chat History ] (optional feature for future)     │
│ [ Delete All ]                                          │
└────────────────────┬────────────────────────────────────┘
                     ↓ User chooses "Delete All"
┌─────────────────────────────────────────────────────────┐
│ POST /api/v1/trips/end                                  │
│ {                                                       │
│   "trip_id": "trip_uuid_123"                            │
│ }                                                       │
└────────────────────┬────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────┐
│ Backend:                                                │
│ 1. Mark trip as ended                                   │
│ 2. Queue all associated DM rooms for deletion           │
│ 3. Leave user from all rooms                            │
│ 4. Schedule room deletion (within 1 hour)               │
└────────────────────┬────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────┐
│ App:                                                    │
│ 1. Delete local message cache                           │
│ 2. Leave all Matrix rooms                               │
│ 3. Stop BLE discovery                                   │
│ 4. Show: "Trip ended. All messages deleted."            │
└─────────────────────────────────────────────────────────┘
```

---

## 🗄 Database Schema

### Trips Table

```sql
CREATE TABLE trips (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  user_id UUID NOT NULL REFERENCES users(id),

  -- Trip details
  route VARCHAR(200) NOT NULL,           -- "Tokyo → Osaka"
  train_number VARCHAR(50),              -- "Nozomi 123" (optional)

  -- Timing
  departure_time TIMESTAMP WITH TIME ZONE NOT NULL,
  estimated_arrival TIMESTAMP WITH TIME ZONE NOT NULL,
  actual_end_time TIMESTAMP WITH TIME ZONE,

  -- Discovery
  discovery_enabled BOOLEAN DEFAULT true,
  ble_anonymous_id VARCHAR(50) NOT NULL, -- "TB_abc123"

  -- Status
  status VARCHAR(20) DEFAULT 'active',   -- active, ended

  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_trips_user_id ON trips(user_id);
CREATE INDEX idx_trips_status ON trips(status);
CREATE INDEX idx_trips_ble_id ON trips(ble_anonymous_id);
```

### Discoveries Table (Anonymized Analytics)

```sql
CREATE TABLE discoveries (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  -- Anonymized (no user_id!)
  trip_route VARCHAR(200) NOT NULL,      -- "Tokyo → Osaka"

  -- Discovery metadata
  discovered_user_anonymous_id VARCHAR(50) NOT NULL,
  distance_estimate VARCHAR(20),         -- "Close (2-10m)"
  rssi INTEGER,                          -- BLE signal strength

  -- Demographics (optional, anonymized)
  discoverer_age_range VARCHAR(20),      -- "25-30"
  discovered_age_range VARCHAR(20),
  discoverer_gender CHAR(1),             -- M/F/O
  discovered_gender CHAR(1),

  discovered_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_discoveries_route ON discoveries(trip_route);
CREATE INDEX idx_discoveries_timestamp ON discoveries(discovered_at);
```

### Matrix Ephemeral Rooms Table

```sql
CREATE TABLE matrix_ephemeral_rooms (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  room_id VARCHAR(255) NOT NULL UNIQUE,  -- !dm_xyz:trainblink.org

  -- Participants (via trips, not direct user_id for privacy)
  trip1_id UUID REFERENCES trips(id),
  trip2_id UUID REFERENCES trips(id),

  -- MLS encryption
  mls_group_id VARCHAR(255),

  -- Ephemeral settings
  expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
  auto_delete_queued BOOLEAN DEFAULT false,
  deleted_at TIMESTAMP WITH TIME ZONE,

  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_ephemeral_rooms_expires ON matrix_ephemeral_rooms(expires_at);
CREATE INDEX idx_ephemeral_rooms_trips ON matrix_ephemeral_rooms(trip1_id, trip2_id);
```

### Matrix Users Table (Existing, Updated)

```sql
CREATE TABLE matrix_users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  user_id UUID NOT NULL UNIQUE REFERENCES users(id),
  matrix_user_id VARCHAR(255) NOT NULL UNIQUE,  -- @trainblink_xxx:trainblink.org

  -- Authentication
  access_token TEXT NOT NULL,                    -- Encrypted
  device_id VARCHAR(255),

  -- Status
  is_active BOOLEAN DEFAULT true,
  last_seen_at TIMESTAMP WITH TIME ZONE,

  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  expires_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() + INTERVAL '30 days'
);
```

---

## 🔌 API Endpoints

### Trip Management API

```http
# Start a trip
POST /api/v1/trips/start
Authorization: Bearer {token}
Content-Type: application/json

Request:
{
  "route": "Tokyo → Osaka",
  "train_number": "Nozomi 123",
  "departure_time": "2025-11-20T10:00:00Z",
  "estimated_arrival": "2025-11-20T12:30:00Z"
}

Response (201):
{
  "status": "success",
  "data": {
    "trip_id": "550e8400-e29b-41d4-a716-446655440000",
    "ble_anonymous_id": "TB_abc123",
    "route": "Tokyo → Osaka",
    "discovery_enabled": true
  }
}

---

# End a trip
POST /api/v1/trips/end
Authorization: Bearer {token}
Content-Type: application/json

Request:
{
  "trip_id": "550e8400-e29b-41d4-a716-446655440000"
}

Response (200):
{
  "status": "success",
  "data": {
    "trip_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "ended",
    "rooms_deleted": 2,
    "messages_deleted": 45
  }
}

---

# Get active trip
GET /api/v1/trips/active
Authorization: Bearer {token}

Response (200):
{
  "status": "success",
  "data": {
    "trip": {
      "trip_id": "550e8400-e29b-41d4-a716-446655440000",
      "route": "Tokyo → Osaka",
      "ble_anonymous_id": "TB_abc123",
      "departure_time": "2025-11-20T10:00:00Z",
      "estimated_arrival": "2025-11-20T12:30:00Z",
      "active_chats": 2
    }
  }
}
```

### Discovery API

```http
# Log discovery (for analytics)
POST /api/v1/discoveries/log
Authorization: Bearer {token}
Content-Type: application/json

Request:
{
  "trip_id": "550e8400-e29b-41d4-a716-446655440000",
  "discovered_anonymous_id": "TB_def456",
  "rssi": -75,
  "distance_estimate": "Close (2-10m)"
}

Response (200):
{
  "status": "success"
}
```

### Matrix DM API

```http
# Create ephemeral DM
POST /api/v1/matrix/dm/create
Authorization: Bearer {token}
Content-Type: application/json

Request:
{
  "with_user_anonymous_id": "TB_def456",
  "trip_id": "550e8400-e29b-41d4-a716-446655440000",
  "enable_mls": true
}

Response (201):
{
  "status": "success",
  "data": {
    "room_id": "!dm_xyz789:trainblink.org",
    "matrix_user_id": "@trainblink_abc123:trainblink.org",
    "access_token": "syt_xxx",
    "device_id": "DEVICE123",
    "homeserver_url": "https://matrix.trainblink.org",
    "mls_group_id": "mls_group_dm_xyz",
    "expires_at": "2025-11-21T10:00:00Z"
  }
}

---

# Delete DM (manual, rare)
DELETE /api/v1/matrix/dm/{room_id}
Authorization: Bearer {token}

Response (200):
{
  "status": "success"
}
```

---

## 📱 Mobile App Implementation Guide

### iOS (Swift) - BLE Broadcasting

```swift
import CoreBluetooth

class BLEDiscoveryManager: NSObject {
    // TrainBlink service UUID (fixed)
    let serviceUUID = CBUUID(string: "TB000000-0000-1000-8000-00805F9B34FB")

    // Current trip info
    var currentTrip: Trip?
    var anonymousID: String = ""  // "TB_abc123"

    // BLE managers
    var peripheralManager: CBPeripheralManager!
    var centralManager: CBCentralManager!

    // Discovered users
    var discoveredUsers: [NearbyUser] = []

    override init() {
        super.init()
        peripheralManager = CBPeripheralManager(delegate: self, queue: nil)
        centralManager = CBCentralManager(delegate: self, queue: nil)
    }

    // Start discovery for a trip
    func startDiscovery(for trip: Trip) {
        self.currentTrip = trip
        self.anonymousID = trip.anonymousID

        // Start broadcasting
        startBroadcasting()

        // Start scanning
        startScanning()
    }

    // BLE Broadcasting
    func startBroadcasting() {
        let advertisementData: [String: Any] = [
            CBAdvertisementDataServiceUUIDsKey: [serviceUUID],
            CBAdvertisementDataLocalNameKey: anonymousID
        ]

        peripheralManager.startAdvertising(advertisementData)
        print("📡 Broadcasting as: \(anonymousID)")
    }

    // BLE Scanning
    func startScanning() {
        centralManager.scanForPeripherals(
            withServices: [serviceUUID],
            options: [CBCentralManagerScanOptionAllowDuplicatesKey: true]
        )
        print("🔍 Scanning for nearby travelers...")
    }

    // Stop discovery
    func stopDiscovery() {
        peripheralManager.stopAdvertising()
        centralManager.stopScan()
        discoveredUsers.removeAll()
        print("🛑 Discovery stopped")
    }
}

// MARK: - CBCentralManagerDelegate

extension BLEDiscoveryManager: CBCentralManagerDelegate {
    func centralManagerDidUpdateState(_ central: CBCentralManager) {
        if central.state == .poweredOn {
            startScanning()
        }
    }

    func centralManager(
        _ central: CBCentralManager,
        didDiscover peripheral: CBPeripheral,
        advertisementData: [String : Any],
        rssi RSSI: NSNumber
    ) {
        // Extract anonymous ID
        guard let localName = advertisementData[CBAdvertisementDataLocalNameKey] as? String,
              localName.hasPrefix("TB_") else {
            return
        }

        // Estimate distance from RSSI
        let distance = estimateDistance(rssi: RSSI.intValue)

        // Filter: only show users within 100m
        if RSSI.intValue < -100 {
            return  // Too far
        }

        // Check if already discovered
        if let index = discoveredUsers.firstIndex(where: { $0.anonymousID == localName }) {
            // Update existing
            discoveredUsers[index].rssi = RSSI.intValue
            discoveredUsers[index].distance = distance
            discoveredUsers[index].lastSeen = Date()
        } else {
            // New user discovered
            let user = NearbyUser(
                anonymousID: localName,
                rssi: RSSI.intValue,
                distance: distance,
                lastSeen: Date()
            )
            discoveredUsers.append(user)

            // Notify UI
            NotificationCenter.default.post(
                name: .didDiscoverNearbyUser,
                object: user
            )

            // Log to backend (analytics)
            logDiscovery(user: user)
        }
    }

    func estimateDistance(rssi: Int) -> String {
        switch rssi {
        case -50...0:
            return "Very Close (0-2m)"
        case -70...-51:
            return "Close (2-10m)"
        case -85...-71:
            return "Medium (10-30m)"
        case -95...-86:
            return "Far (30-50m)"
        case -100...-96:
            return "Very Far (50-100m)"
        default:
            return "Out of Range"
        }
    }

    func logDiscovery(user: NearbyUser) {
        guard let tripID = currentTrip?.id else { return }

        Task {
            try? await apiClient.logDiscovery(
                tripID: tripID,
                anonymousID: user.anonymousID,
                rssi: user.rssi,
                distance: user.distance
            )
        }
    }
}

// MARK: - Models

struct Trip {
    let id: UUID
    let route: String
    let anonymousID: String  // "TB_abc123"
    let departureTime: Date
    let estimatedArrival: Date
}

struct NearbyUser {
    let anonymousID: String
    var rssi: Int
    var distance: String
    var lastSeen: Date
}
```

---

## ⏱ Auto-Cleanup Service

### Backend Cron Job (Go)

```go
// Auto-delete expired ephemeral rooms
func (s *CleanupService) DeleteExpiredRooms() error {
    ctx := context.Background()

    // Find expired rooms
    var expiredRooms []MatrixEphemeralRoom
    err := s.db.WithContext(ctx).
        Where("expires_at < NOW()").
        Where("deleted_at IS NULL").
        Find(&expiredRooms).Error

    if err != nil {
        return err
    }

    for _, room := range expiredRooms {
        // Delete from Matrix Synapse
        if err := s.matrixClient.DeleteRoom(ctx, room.RoomID); err != nil {
            log.Printf("Failed to delete room %s: %v", room.RoomID, err)
            continue
        }

        // Mark as deleted
        s.db.Model(&room).Updates(map[string]interface{}{
            "deleted_at": time.Now(),
        })

        log.Printf("✅ Deleted expired room: %s", room.RoomID)
    }

    return nil
}

// Run every 1 hour
func StartCleanupService(db *gorm.DB, matrixClient *matrix.Client) {
    service := &CleanupService{db: db, matrixClient: matrixClient}

    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()

    for range ticker.C {
        if err := service.DeleteExpiredRooms(); err != nil {
            log.Printf("❌ Cleanup failed: %v", err)
        }
    }
}
```

---

## 🎯 MVP Success Criteria

### Functional Requirements
- [ ] User can manually start a trip
- [ ] BLE discovers users within 50-100m
- [ ] User can initiate 1-on-1 chat
- [ ] iOS ↔ Android messaging works via Matrix
- [ ] Messages are E2EE with MLS
- [ ] Messages auto-delete when trip ends
- [ ] User can manually end trip

### Non-Functional Requirements
- [ ] BLE discovery: <100ms latency
- [ ] Message delivery: <2s (when online)
- [ ] Privacy: Server can't read messages
- [ ] Battery: <5% drain per hour
- [ ] Offline: AirDrop/Nearby Share works

### Analytics
- [ ] Track discoveries per route (anonymized)
- [ ] Track chat initiation rate
- [ ] Track average trip duration
- [ ] Track message count per trip

---

## 🚀 Implementation Order

### Week 1: Database & Core Services
1. ✅ Create database migrations (trips, discoveries, ephemeral_rooms)
2. ✅ Implement Trip Management Service
3. ✅ Implement Discovery Tracking Service
4. ✅ Update Matrix User Manager for ephemeral users

### Week 2: Matrix Integration
5. ✅ Update Matrix Room Manager for ephemeral DMs
6. ✅ Integrate MLS with ephemeral rooms
7. ✅ Implement auto-cleanup service
8. ✅ API endpoints (trips, discoveries, matrix/dm)

### Week 3: Mobile App (BLE)
9. ✅ iOS BLE broadcasting/scanning
10. ✅ Android BLE implementation
11. ✅ Distance estimation logic
12. ✅ UI for discovered users

### Week 4: Matrix SDK Integration
13. ✅ iOS Matrix SDK setup
14. ✅ Android Matrix SDK setup
15. ✅ Chat interface
16. ✅ Local caching during trip

### Week 5: Testing & Polish
17. ✅ Integration tests
18. ✅ E2E testing (iOS ↔ Android)
19. ✅ Battery optimization
20. ✅ Privacy audit

---

**Ready to implement! Shall I start with Week 1 (Database & Core Services)?** 🚀
