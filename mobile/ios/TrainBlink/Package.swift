// swift-tools-version:5.9
import PackageDescription

let package = Package(
    name: "TrainBlink",
    platforms: [
        .iOS(.v16)
    ],
    products: [
        .library(
            name: "TrainBlink",
            targets: ["TrainBlink"]
        )
    ],
    dependencies: [
        // Matrix SDK
        .package(url: "https://github.com/matrix-org/matrix-ios-sdk.git", from: "0.27.0"),
        // Networking
        .package(url: "https://github.com/Alamofire/Alamofire.git", from: "5.8.0")
    ],
    targets: [
        .target(
            name: "TrainBlink",
            dependencies: [
                .product(name: "MatrixSDK", package: "matrix-ios-sdk"),
                "Alamofire"
            ],
            path: "TrainBlink"
        ),
        .testTarget(
            name: "TrainBlinkTests",
            dependencies: ["TrainBlink"]
        )
    ]
)
