# TrainBlink Mobile App - Week 3 Implementation

## Overview

This directory contains the iOS and Android mobile applications for TrainBlink, implementing **Week 3: BLE Discovery Layer** from the approved implementation plan.

## Project Structure

```
mobile/
├── ios/
│   └── TrainBlink/
│       ├── Package.swift                 # Swift Package Manager configuration
│       └── TrainBlink/
│           ├── Models/
│           │   ├── Trip.swift           # Trip data models
│           │   └── Discovery.swift      # Discovery & distance estimation
│           ├── Services/
│           │   ├── BLEDiscoveryManager.swift  # CoreBluetooth BLE manager
│           │   └── APIService.swift           # Backend HTTP client
│           ├── Views/
│           │   ├── TripSelectionView.swift    # Trip selection UI
│           │   └── NearbyTravelersView.swift  # Nearby travelers UI
│           ├── Config/
│           │   └── AppConfig.swift            # App configuration
│           └── Info.plist                     # iOS permissions & config
│
└── android/
    └── TrainBlink/
        ├── build.gradle                  # Project-level Gradle config
        └── app/
            ├── build.gradle              # App-level Gradle config
            ├── src/main/
            │   ├── AndroidManifest.xml   # Android permissions & config
            │   └── java/org/trainblink/app/
            │       ├── models/
            │       │   ├── Trip.kt       # Trip data models
            │       │   └── Discovery.kt  # Discovery & distance estimation
            │       ├── services/
            │       │   └── BLEDiscoveryService.kt  # BLE foreground service
            │       └── network/
            │           └── ApiService.kt           # Retrofit HTTP client
```

## Implementation Details

### iOS (Swift + CoreBluetooth)

#### BLEDiscoveryManager.swift (393 lines)
- **Central Manager**: Scans for nearby TrainBlink users
- **Peripheral Manager**: Broadcasts anonymous BLE ID
- **Service UUID**: `TB000000-0000-1000-8000-00805F9B34FB`
- **Characteristic UUID**: `TB000001-0000-1000-8000-00805F9B34FB`
- **RSSI Distance Estimation**: 5 categories from Very Close (0-2m) to Very Far (50-100m)
- **Automatic Cleanup**: Removes stale discoveries (>30s) every 10 seconds
- **Analytics Integration**: Logs discovery events to backend API

**Key Features:**
- Dual-mode operation (Central + Peripheral simultaneously)
- Real-time discovery updates via Combine `@Published` properties
- Background modes support for continuous discovery
- Automatic reconnection handling

#### TripSelectionView.swift (439 lines)
- **SwiftUI-based** trip selection and management
- Popular routes picker (Tokyo → Osaka, Paris → Lyon, etc.)
- Active trip display with real-time countdown
- Discovery toggle with live BLE state
- Navigation to Nearby Travelers view

#### NearbyTravelersView.swift (320 lines)
- Real-time nearby travelers list
- Distance-based color coding (green = very close, red = very far)
- Signal strength (RSSI) display
- "Active" indicator for fresh discoveries (<30s)
- Create chat modal with privacy features

### Android (Kotlin + BluetoothLE)

#### BLEDiscoveryService.kt (422 lines)
- **Foreground Service**: Runs in background with persistent notification
- **GATT Server**: Advertises anonymous BLE ID
- **BLE Scanner**: Discovers nearby users
- **StateFlow Integration**: Reactive state management
- **Coroutine-based**: Async API calls with proper scope management

**Key Features:**
- Foreground service for reliable background operation
- Android 12+ permission handling (BLUETOOTH_SCAN, BLUETOOTH_ADVERTISE)
- GATT server for characteristic reads
- Automatic GATT connection management
- Stale discovery cleanup (>30s)

#### ApiService.kt (205 lines)
- **Retrofit 2.9.0** HTTP client
- **Gson** JSON serialization with date format handling
- **OkHttp** with logging interceptor
- **Authentication**: X-User-ID (dev) / JWT (production)
- All 15 backend endpoints (trips, discoveries, Matrix DMs)

## RSSI Distance Estimation

Both platforms use identical distance estimation logic:

| RSSI Range | Distance | Category | Color |
|------------|----------|----------|-------|
| -50 to 0 | 0-2m | Very Close | Green |
| -70 to -51 | 2-10m | Close | Light Green |
| -85 to -71 | 10-30m | Medium | Yellow |
| -95 to -86 | 30-50m | Far | Orange |
| < -95 | 50-100m | Very Far | Red |

## BLE Protocol

### Service UUID
```
TB000000-0000-1000-8000-00805F9B34FB
```

### Characteristic UUID
```
TB000001-0000-1000-8000-00805F9B34FB
```

### Broadcasting
- **Properties**: READ, NOTIFY
- **Value**: Anonymous BLE ID (e.g., "TB_abc123")
- **Encoding**: UTF-8

### Discovery Flow
1. **Broadcast**: Advertise service UUID + characteristic with anonymous ID
2. **Scan**: Look for devices advertising TrainBlink service
3. **Connect**: Establish GATT connection to discovered peripheral
4. **Read**: Read characteristic value (anonymous ID)
5. **Process**: Calculate distance from RSSI, update UI, log to backend
6. **Disconnect**: Close GATT connection after reading

## Permissions

### iOS (Info.plist)

```xml
<key>NSBluetoothAlwaysUsageDescription</key>
<string>TrainBlink uses Bluetooth to discover nearby travelers...</string>

<key>NSLocationWhenInUseUsageDescription</key>
<string>Location permission required for Bluetooth (not tracked)</string>

<key>UIBackgroundModes</key>
<array>
    <string>bluetooth-central</string>
    <string>bluetooth-peripheral</string>
</array>
```

### Android (AndroidManifest.xml)

```xml
<!-- Android 12+ -->
<uses-permission android:name="android.permission.BLUETOOTH_SCAN" />
<uses-permission android:name="android.permission.BLUETOOTH_ADVERTISE" />
<uses-permission android:name="android.permission.BLUETOOTH_CONNECT" />

<!-- Location (required for BLE) -->
<uses-permission android:name="android.permission.ACCESS_FINE_LOCATION" />

<!-- Foreground Service -->
<uses-permission android:name="android.permission.FOREGROUND_SERVICE" />
<uses-permission android:name="android.permission.FOREGROUND_SERVICE_CONNECTED_DEVICE" />
```

## API Integration

### Backend Endpoints Used

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/v1/trips/start` | POST | Start new trip |
| `/api/v1/trips/end` | POST | End active trip |
| `/api/v1/trips/active` | GET | Get user's active trip |
| `/api/v1/trips/{id}/discovery` | PATCH | Toggle discovery |
| `/api/v1/discoveries/log` | POST | Log discovery event |
| `/api/v1/matrix/dm/create` | POST | Create ephemeral DM |

### Authentication

**Development:**
```
X-User-ID: <uuid>
```

**Production:**
```
Authorization: Bearer <jwt_token>
```

## Dependencies

### iOS (Swift Package Manager)

```swift
dependencies: [
    .package(url: "https://github.com/matrix-org/matrix-ios-sdk.git", from: "0.27.0"),
    .package(url: "https://github.com/Alamofire/Alamofire.git", from: "5.8.0")
]
```

### Android (Gradle)

```gradle
// Compose
implementation "androidx.compose.ui:ui:1.5.4"
implementation "androidx.compose.material3:material3:1.1.2"

// Networking
implementation 'com.squareup.retrofit2:retrofit:2.9.0'
implementation 'com.squareup.retrofit2:converter-gson:2.9.0'

// Matrix SDK
implementation 'org.matrix.android:matrix-android-sdk2:1.6.8'

// Bluetooth LE
implementation 'androidx.bluetooth:bluetooth:1.0.0-alpha01'
```

## Building & Running

### iOS

```bash
cd mobile/ios/TrainBlink
swift build
# Or open in Xcode:
open TrainBlink.xcodeproj
```

**Requirements:**
- Xcode 15.0+
- iOS 16.0+
- Swift 5.9+

### Android

```bash
cd mobile/android/TrainBlink
./gradlew assembleDebug
# Or open in Android Studio
```

**Requirements:**
- Android Studio Hedgehog (2023.1.1)+
- Android SDK 26+ (minSdk)
- Target SDK 34
- Kotlin 1.9.20

## Testing

### Manual Testing Checklist

#### BLE Discovery
- [ ] Start trip and verify BLE broadcast starts
- [ ] Scan for nearby devices
- [ ] Verify distance estimation accuracy
- [ ] Check stale discovery cleanup (30s timeout)
- [ ] Test background operation
- [ ] Verify permissions are requested

#### Trip Management
- [ ] Create new trip with route
- [ ] View active trip details
- [ ] Toggle discovery on/off
- [ ] End trip and verify BLE stops
- [ ] Check trip history

#### Analytics
- [ ] Verify discovery events logged to backend
- [ ] Check RSSI values are correct
- [ ] Confirm no PII in discovery logs

### Device Testing

**Minimum:**
- iOS: iPhone 8 or newer (iOS 16+)
- Android: Any device with BLE support (Android 8.0+)

**Recommended:**
- Test on two physical devices simultaneously
- Verify cross-platform (iOS ↔ Android)
- Test in various environments (train, crowded platform)

## Privacy & Security

### Anonymous Discovery
- BLE broadcasts only anonymous ID (TB_xxxxx)
- No names, phone numbers, or user IDs transmitted
- Discovery events log route + anonymous IDs only
- RSSI values provide coarse proximity, not precise location

### Data Retention
- Fresh discoveries: Active for 30 seconds
- Stale discoveries: Automatically removed
- Backend logs: Anonymized, no user_id in discoveries table

## Known Limitations

1. **iOS Background Scanning**: May be throttled by iOS after ~3 minutes in background
2. **Android Permissions**: Requires location permission (Android system requirement for BLE)
3. **RSSI Accuracy**: Varies by device, obstacles, and interference
4. **Range**: Effective range ~50-100m, but can vary significantly
5. **Cross-Platform**: Matrix SDK integration pending (Week 4)

## Next Steps: Week 4

Week 4 will implement **Matrix SDK Integration**:

### iOS
- Integrate matrix-ios-sdk
- MatrixChatManager service
- Chat UI with message bubbles
- Room lifecycle management
- Offline message queuing

### Android
- Integrate matrix-android-sdk2
- MatrixChatService
- Jetpack Compose chat UI
- Room expiry countdown
- Message persistence

## Troubleshooting

### iOS

**BLE not advertising:**
- Check Bluetooth permission in Settings
- Verify Info.plist has correct usage descriptions
- Ensure app has location permission (iOS requirement)

**Scanning not finding devices:**
- Check Background Modes are enabled in project capabilities
- Verify service UUID matches exactly
- Restart Bluetooth (toggle off/on)

### Android

**Service crashes:**
- Check all Bluetooth permissions granted
- Verify foreground service notification created
- Review logcat for detailed error messages

**Discovery not working:**
- Ensure location services enabled (Android requirement)
- Check BLUETOOTH_SCAN permission (Android 12+)
- Verify GATT server started successfully

## Contributing

When adding features:
1. Follow existing code structure and naming conventions
2. Update this README with new components
3. Add inline documentation for complex logic
4. Test on both physical iOS and Android devices
5. Ensure privacy requirements maintained

## License

Copyright (c) 2025 TrainBlink. All rights reserved.
