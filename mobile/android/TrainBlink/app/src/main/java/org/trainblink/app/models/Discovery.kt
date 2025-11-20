package org.trainblink.app.models

import android.os.Parcelable
import com.google.gson.annotations.SerializedName
import kotlinx.parcelize.Parcelize
import java.util.Date

/**
 * Distance estimation based on RSSI signal strength
 */
enum class DistanceEstimate(val rawValue: String) {
    VERY_CLOSE("Very Close (0-2m)"),
    CLOSE("Close (2-10m)"),
    MEDIUM("Medium (10-30m)"),
    FAR("Far (30-50m)"),
    VERY_FAR("Very Far (50-100m)");

    companion object {
        /**
         * Initialize from RSSI value
         */
        fun fromRssi(rssi: Int): DistanceEstimate {
            return when {
                rssi >= -50 -> VERY_CLOSE
                rssi >= -70 -> CLOSE
                rssi >= -85 -> MEDIUM
                rssi >= -95 -> FAR
                else -> VERY_FAR
            }
        }
    }

    /**
     * Get emoji icon for distance
     */
    val icon: String
        get() = when (this) {
            VERY_CLOSE -> "👤"
            CLOSE -> "🚶"
            MEDIUM -> "🚶‍♂️"
            FAR -> "🏃"
            VERY_FAR -> "🏃‍♂️"
        }

    /**
     * Get color for distance indicator (Material colors)
     */
    val colorHex: String
        get() = when (this) {
            VERY_CLOSE -> "#4CAF50"  // Green
            CLOSE -> "#8BC34A"       // Light Green
            MEDIUM -> "#FFEB3B"      // Yellow
            FAR -> "#FF9800"         // Orange
            VERY_FAR -> "#F44336"    // Red
        }

    /**
     * Get short description
     */
    val shortDescription: String
        get() = when (this) {
            VERY_CLOSE -> "Very Close"
            CLOSE -> "Close"
            MEDIUM -> "Medium"
            FAR -> "Far"
            VERY_FAR -> "Very Far"
        }
}

/**
 * Discovered user via BLE
 */
@Parcelize
data class DiscoveredUser(
    val id: String, // BLE Anonymous ID (TB_xxxxx)
    var rssi: Int,
    var distance: DistanceEstimate,
    val discoveredAt: Date,
    var lastSeen: Date
) : Parcelable {

    /**
     * Check if discovery is fresh (within last 30 seconds)
     */
    val isFresh: Boolean
        get() = (Date().time - lastSeen.time) < 30_000

    /**
     * Time since last seen in milliseconds
     */
    val timeSinceLastSeen: Long
        get() = Date().time - lastSeen.time

    /**
     * Update with new RSSI reading
     */
    fun update(newRssi: Int) {
        rssi = newRssi
        distance = DistanceEstimate.fromRssi(newRssi)
        lastSeen = Date()
    }
}

/**
 * Request payload for logging a discovery event
 */
data class LogDiscoveryRequest(
    @SerializedName("trip_route")
    val tripRoute: String,
    @SerializedName("discovered_user_anonymous_id")
    val discoveredUserAnonymousId: String,
    @SerializedName("distance_estimate")
    val distanceEstimate: String,
    val rssi: Int,
    @SerializedName("discoverer_age_range")
    val discovererAgeRange: String? = null,
    @SerializedName("discovered_age_range")
    val discoveredAgeRange: String? = null,
    @SerializedName("discoverer_gender")
    val discovererGender: String? = null,
    @SerializedName("discovered_gender")
    val discoveredGender: String? = null
)

/**
 * Discovery analytics response
 */
data class DiscoveryStatsResponse(
    val status: String,
    val data: DiscoveryStats?
) {
    data class DiscoveryStats(
        @SerializedName("total_discoveries")
        val totalDiscoveries: Int,
        @SerializedName("unique_users")
        val uniqueUsers: Int,
        @SerializedName("average_rssi")
        val averageRssi: Double,
        @SerializedName("most_common_distance")
        val mostCommonDistance: String
    )
}
