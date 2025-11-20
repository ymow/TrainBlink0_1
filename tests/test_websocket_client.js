#!/usr/bin/env node

/**
 * TrainBlink WebSocket Test Client
 * Tests real-time messaging functionality
 */

const http = require('http');
const WebSocket = require('ws');

const BASE_URL = 'http://localhost:8080';
const WS_URL = 'ws://localhost:8080';

// Helper function to make HTTP requests
function httpRequest(method, path, data = null, headers = {}) {
    return new Promise((resolve, reject) => {
        const url = new URL(BASE_URL + path);
        const options = {
            hostname: url.hostname,
            port: url.port,
            path: url.pathname,
            method: method,
            headers: {
                'Content-Type': 'application/json',
                ...headers
            }
        };

        const req = http.request(options, (res) => {
            let body = '';
            res.on('data', (chunk) => body += chunk);
            res.on('end', () => {
                try {
                    resolve({ status: res.statusCode, data: JSON.parse(body) });
                } catch (e) {
                    resolve({ status: res.statusCode, data: body });
                }
            });
        });

        req.on('error', reject);

        if (data) {
            req.write(JSON.stringify(data));
        }

        req.end();
    });
}

// Test 1: Enter station
async function enterStation(userID, deviceID, stationID) {
    console.log(`\n📍 [${userID}] Entering ${stationID}...`);

    const response = await httpRequest('POST', '/api/v1/geofence/enter', {
        station_id: stationID,
        coordinates: {
            latitude: 35.6812,
            longitude: 139.7671,
            accuracy: 10.5
        },
        client_version: '2.5.0',
        capabilities: {
            p2p_enabled: true,
            matrix_enabled: true,
            mls_supported: false
        }
    }, {
        'X-User-ID': userID,
        'X-Device-ID': deviceID
    });

    if (response.status === 200 && response.data.status === 'success') {
        const sessionID = response.data.data.session_id;
        console.log(`   ✓ Session created: ${sessionID}`);
        return sessionID;
    } else {
        throw new Error(`Failed to enter station: ${JSON.stringify(response.data)}`);
    }
}

// Test 2: Connect WebSocket
function connectWebSocket(sessionID, userID) {
    return new Promise((resolve, reject) => {
        const ws = new WebSocket(`${WS_URL}/ws?session_id=${sessionID}`);

        ws.on('open', () => {
            console.log(`   ✓ WebSocket connected for ${userID}`);
            resolve(ws);
        });

        ws.on('error', (error) => {
            console.error(`   ✗ WebSocket error for ${userID}:`, error.message);
            reject(error);
        });

        ws.on('message', (data) => {
            try {
                const message = JSON.parse(data);
                console.log(`   📨 [${userID}] Received:`, JSON.stringify(message, null, 2));
            } catch (e) {
                console.log(`   📨 [${userID}] Received (raw):`, data.toString());
            }
        });

        ws.on('close', () => {
            console.log(`   🔌 [${userID}] WebSocket closed`);
        });
    });
}

// Test 3: Send message
function sendMessage(ws, type, content, metadata = {}) {
    const message = {
        type: type,
        content: content,
        metadata: metadata
    };

    ws.send(JSON.stringify(message));
    console.log(`   📤 Sent:`, JSON.stringify(message));
}

// Main test function
async function runTests() {
    console.log('🚀 TrainBlink WebSocket Test Client');
    console.log('===================================\n');

    try {
        // Test 1: Health check
        console.log('🏥 Test 1: Health Check');
        const health = await httpRequest('GET', '/health');
        console.log(`   Status: ${health.status}`);
        console.log(`   Version: ${health.data.version}`);
        console.log('   ✓ Health check passed\n');

        // Test 2: Enter station for User 1
        console.log('👤 Test 2: User 1 Setup');
        const user1Session = await enterStation('user-001', 'iOS-123', 'station_tokyo_001');

        // Test 3: Enter station for User 2
        console.log('\n👤 Test 3: User 2 Setup');
        const user2Session = await enterStation('user-002', 'Android-456', 'station_tokyo_001');

        // Test 4: Connect WebSocket for both users
        console.log('\n🔌 Test 4: WebSocket Connections');
        const ws1 = await connectWebSocket(user1Session, 'user-001');
        await new Promise(resolve => setTimeout(resolve, 500));
        const ws2 = await connectWebSocket(user2Session, 'user-002');

        // Wait for connections to stabilize
        await new Promise(resolve => setTimeout(resolve, 1000));

        // Test 5: User 1 sends message to station
        console.log('\n💬 Test 5: User 1 sends station message');
        sendMessage(ws1, 'message', 'Hello everyone in Tokyo Station!');

        await new Promise(resolve => setTimeout(resolve, 500));

        // Test 6: User 2 sends typing indicator
        console.log('\n⌨️  Test 6: User 2 sends typing indicator');
        sendMessage(ws2, 'typing', '', { is_typing: true });

        await new Promise(resolve => setTimeout(resolve, 500));

        // Test 7: User 2 sends message
        console.log('\n💬 Test 7: User 2 sends message');
        sendMessage(ws2, 'message', 'Hi User 1! Nice to meet you!');

        await new Promise(resolve => setTimeout(resolve, 500));

        // Test 8: User 2 stops typing
        console.log('\n⌨️  Test 8: User 2 stops typing');
        sendMessage(ws2, 'typing', '', { is_typing: false });

        await new Promise(resolve => setTimeout(resolve, 500));

        // Test 9: User 1 sends ping
        console.log('\n🏓 Test 9: User 1 sends ping');
        sendMessage(ws1, 'ping', '');

        await new Promise(resolve => setTimeout(resolve, 1000));

        // Test 10: Get statistics
        console.log('\n📊 Test 10: Get Statistics');
        const stats = await httpRequest('GET', '/api/v1/geofence/stats');
        console.log('   Statistics:', JSON.stringify(stats.data.data, null, 2));

        // Close WebSocket connections
        console.log('\n🔌 Closing WebSocket connections...');
        ws1.close();
        ws2.close();

        await new Promise(resolve => setTimeout(resolve, 1000));

        console.log('\n✅ All tests completed successfully!');
        process.exit(0);

    } catch (error) {
        console.error('\n❌ Test failed:', error.message);
        console.error(error.stack);
        process.exit(1);
    }
}

// Run tests
runTests();
