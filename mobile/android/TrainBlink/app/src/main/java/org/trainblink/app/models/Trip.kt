package org.trainblink.app.models

import android.os.Parcelable
import com.google.gson.annotations.SerializedName
import kotlinx.parcelize.Parcelize
import java.util.Date
import java.util.UUID

/**
 * Trip status enumeration
 */
enum class TripStatus {
    @SerializedName("active")
    ACTIVE,

    @SerializedName("ended")
    ENDED,

    @SerializedName("cancelled")
    CANCELLED
}

/**
 * Trip model representing a user's journey
 */
@Parcelize
data class Trip(
    val id: UUID,
    @SerializedName("user_id")
    val userId: UUID,
    val route: String,
    @SerializedName("train_number")
    val trainNumber: String?,
    @SerializedName("departure_time")
    val departureTime: Date,
    @SerializedName("estimated_arrival")
    val estimatedArrival: Date,
    @SerializedName("actual_end_time")
    val actualEndTime: Date?,
    @SerializedName("discovery_enabled")
    val discoveryEnabled: Boolean,
    @SerializedName("ble_anonymous_id")
    val bleAnonymousId: String,
    val status: TripStatus,
    @SerializedName("created_at")
    val createdAt: Date,
    @SerializedName("updated_at")
    val updatedAt: Date
) : Parcelable {

    /**
     * Check if trip is currently active
     */
    val isActive: Boolean
        get() = status == TripStatus.ACTIVE && Date().before(estimatedArrival)

    /**
     * Calculate trip duration in milliseconds
     */
    val duration: Long
        get() = estimatedArrival.time - departureTime.time

    /**
     * Calculate remaining time in milliseconds
     */
    val remainingTime: Long
        get() = maxOf(0, estimatedArrival.time - Date().time)

    /**
     * Format remaining time as string
     */
    val remainingTimeFormatted: String
        get() {
            val remaining = remainingTime
            val hours = (remaining / (1000 * 60 * 60)).toInt()
            val minutes = ((remaining / (1000 * 60)) % 60).toInt()

            return if (hours > 0) {
                "${hours}h ${minutes}m remaining"
            } else {
                "${minutes}m remaining"
            }
        }
}

/**
 * Request payload for starting a new trip
 */
data class StartTripRequest(
    val route: String,
    @SerializedName("train_number")
    val trainNumber: String?,
    @SerializedName("departure_time")
    val departureTime: Date,
    @SerializedName("estimated_arrival")
    val estimatedArrival: Date
)

/**
 * Request payload for ending a trip
 */
data class EndTripRequest(
    @SerializedName("trip_id")
    val tripId: UUID
)

/**
 * API response wrapper for single trip
 */
data class TripResponse(
    val status: String,
    val data: Trip?,
    val error: String?
)

/**
 * API response wrapper for trip list
 */
data class TripsListResponse(
    val status: String,
    val data: TripsData?,
    val error: String?
) {
    data class TripsData(
        val trips: List<Trip>,
        val total: Int,
        val limit: Int,
        val offset: Int
    )
}

/**
 * Update discovery enabled request
 */
data class UpdateDiscoveryRequest(
    @SerializedName("discovery_enabled")
    val discoveryEnabled: Boolean
)
