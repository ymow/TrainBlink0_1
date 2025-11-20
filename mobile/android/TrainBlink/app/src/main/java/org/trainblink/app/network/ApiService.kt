package org.trainblink.app.network

import android.content.Context
import com.google.gson.GsonBuilder
import okhttp3.Interceptor
import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import org.trainblink.app.BuildConfig
import org.trainblink.app.models.*
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory
import retrofit2.http.*
import java.util.UUID
import java.util.concurrent.TimeUnit

/**
 * API Service interface for backend communication
 */
interface ApiService {

    // MARK: - Trip Endpoints

    @POST("/api/v1/trips/start")
    suspend fun startTrip(@Body request: StartTripRequest): TripResponse

    @POST("/api/v1/trips/end")
    suspend fun endTrip(@Body request: EndTripRequest): TripResponse

    @GET("/api/v1/trips/active")
    suspend fun getActiveTrip(): TripResponse

    @GET("/api/v1/trips")
    suspend fun getUserTrips(
        @Query("limit") limit: Int = 10,
        @Query("offset") offset: Int = 0
    ): TripsListResponse

    @DELETE("/api/v1/trips/{id}")
    suspend fun cancelTrip(@Path("id") tripId: UUID): TripResponse

    @PATCH("/api/v1/trips/{id}/discovery")
    suspend fun updateDiscoveryEnabled(
        @Path("id") tripId: UUID,
        @Body request: UpdateDiscoveryRequest
    ): TripResponse

    // MARK: - Discovery Endpoints

    @POST("/api/v1/discoveries/log")
    suspend fun logDiscovery(@Body request: LogDiscoveryRequest)

    @GET("/api/v1/discoveries/stats")
    suspend fun getDiscoveryStats(): DiscoveryStatsResponse

    // MARK: - Matrix Endpoints

    @POST("/api/v1/matrix/dm/create")
    suspend fun createEphemeralDM(@Body request: CreateEphemeralDMRequest): MatrixRoomResponse

    @GET("/api/v1/matrix/dm/active")
    suspend fun getActiveEphemeralDMs(): MatrixRoomsResponse

    @PATCH("/api/v1/matrix/dm/{id}/extend")
    suspend fun extendRoomLifetime(
        @Path("id") roomId: UUID,
        @Body request: ExtendRoomRequest
    ): Map<String, Any>

    companion object {
        /**
         * Create API service instance
         */
        fun create(context: Context, userId: UUID? = null, jwtToken: String? = null): ApiService {
            val gson = GsonBuilder()
                .setDateFormat("yyyy-MM-dd'T'HH:mm:ss.SSS'Z'")
                .create()

            // Create authentication interceptor
            val authInterceptor = Interceptor { chain ->
                val original = chain.request()
                val requestBuilder = original.newBuilder()

                // Add user ID header (development)
                userId?.let {
                    requestBuilder.header("X-User-ID", it.toString())
                }

                // Add JWT token (production)
                jwtToken?.let {
                    requestBuilder.header("Authorization", "Bearer $it")
                }

                val request = requestBuilder.build()
                chain.proceed(request)
            }

            // Create logging interceptor
            val loggingInterceptor = HttpLoggingInterceptor().apply {
                level = if (BuildConfig.DEBUG) {
                    HttpLoggingInterceptor.Level.BODY
                } else {
                    HttpLoggingInterceptor.Level.NONE
                }
            }

            // Create OkHttp client
            val okHttpClient = OkHttpClient.Builder()
                .addInterceptor(authInterceptor)
                .addInterceptor(loggingInterceptor)
                .connectTimeout(30, TimeUnit.SECONDS)
                .readTimeout(30, TimeUnit.SECONDS)
                .writeTimeout(30, TimeUnit.SECONDS)
                .build()

            // Create Retrofit instance
            val retrofit = Retrofit.Builder()
                .baseUrl(BuildConfig.API_BASE_URL)
                .client(okHttpClient)
                .addConverterFactory(GsonConverterFactory.create(gson))
                .build()

            return retrofit.create(ApiService::class.java)
        }
    }
}

// MARK: - Matrix Request/Response Models

/**
 * Request to create ephemeral DM
 */
data class CreateEphemeralDMRequest(
    @SerializedName("discovered_user_ble_id")
    val discoveredUserBLEID: String,
    @SerializedName("mls_group_id")
    val mlsGroupId: String? = null
)

/**
 * Request to extend room lifetime
 */
data class ExtendRoomRequest(
    @SerializedName("extension_hours")
    val extensionHours: Int
)

/**
 * Matrix ephemeral room model
 */
data class MatrixEphemeralRoom(
    val id: UUID,
    @SerializedName("room_id")
    val roomId: String,
    @SerializedName("trip1_id")
    val trip1Id: UUID,
    @SerializedName("trip2_id")
    val trip2Id: UUID,
    @SerializedName("anonymous_id_1")
    val anonymousId1: String,
    @SerializedName("anonymous_id_2")
    val anonymousId2: String,
    @SerializedName("mls_group_id")
    val mlsGroupId: String?,
    @SerializedName("expires_at")
    val expiresAt: java.util.Date,
    @SerializedName("message_count")
    val messageCount: Int,
    @SerializedName("last_message_at")
    val lastMessageAt: java.util.Date?,
    @SerializedName("created_at")
    val createdAt: java.util.Date
) {
    /**
     * Time remaining until expiry in milliseconds
     */
    val timeRemaining: Long
        get() = maxOf(0, expiresAt.time - java.util.Date().time)

    /**
     * Check if room is expiring soon (within 1 hour)
     */
    val isExpiringSoon: Boolean
        get() = timeRemaining < 3600_000

    /**
     * Format time remaining
     */
    val timeRemainingFormatted: String
        get() {
            val hours = (timeRemaining / (1000 * 60 * 60)).toInt()
            val minutes = ((timeRemaining / (1000 * 60)) % 60).toInt()

            return when {
                hours > 0 -> "${hours}h ${minutes}m"
                minutes > 0 -> "${minutes}m"
                else -> "Expiring soon"
            }
        }
}

/**
 * API response for single Matrix room
 */
data class MatrixRoomResponse(
    val status: String,
    val data: MatrixEphemeralRoom?
)

/**
 * API response for Matrix room list
 */
data class MatrixRoomsResponse(
    val status: String,
    val data: RoomsData?
) {
    data class RoomsData(
        val rooms: List<MatrixEphemeralRoom>,
        val total: Int
    )
}
