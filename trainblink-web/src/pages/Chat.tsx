import { useState } from 'react';
import { motion } from 'framer-motion';
import { useWebSocket } from '../hooks/useWebSocket';
import { useConnections, useSendMessage } from '../hooks/useStats';

const WEBSOCKET_URL = import.meta.env.VITE_API_URL?.replace('http', 'ws') + '/ws' || 'ws://localhost:8080/ws';

export default function Chat() {
  const [messageInput, setMessageInput] = useState('');
  const [username, setUsername] = useState('');
  const [isUsernameSet, setIsUsernameSet] = useState(false);
  const [usePostApi, setUsePostApi] = useState(false);
  
  const { messages, isConnected, error, sendMessage } = useWebSocket(WEBSOCKET_URL);
  const { data: connections } = useConnections();
  const sendMessageMutation = useSendMessage();

  const handleSendMessage = (e: React.FormEvent) => {
    e.preventDefault();
    if (messageInput.trim()) {
      const fullMessage = `[${username || 'Anonymous'}]: ${messageInput.trim()}`;
      
      if (usePostApi) {
        // Send via POST API
        sendMessageMutation.mutate({
          message: fullMessage,
          type: 'chat'
        });
      } else if (isConnected) {
        // Send via WebSocket
        sendMessage(fullMessage);
      }
      setMessageInput('');
    }
  };

  const handleSetUsername = (e: React.FormEvent) => {
    e.preventDefault();
    if (username.trim()) {
      setIsUsernameSet(true);
    }
  };

  if (!isUsernameSet) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center px-4">
        <motion.div
          className="glass rounded-lg p-8 w-full max-w-md"
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8 }}
        >
          <div className="text-center mb-6">
            <div className="text-4xl mb-4">🚂</div>
            <h1 className="text-2xl font-bold text-white mb-2">Join TrainBlink Chat</h1>
            <p className="text-gray-300">Enter a username to start chatting</p>
          </div>
          
          <form onSubmit={handleSetUsername}>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="Your username..."
              className="w-full px-4 py-3 rounded-lg bg-white/10 border border-white/20 text-white placeholder-gray-400 focus:outline-none focus:border-blue-400 mb-4"
              maxLength={20}
              required
            />
            <button
              type="submit"
              className="w-full button-primary py-3 text-lg"
              disabled={!username.trim()}
            >
              Start Chatting 💬
            </button>
          </form>
          
          <div className="mt-4 text-center">
            <button
              className="text-gray-400 hover:text-white transition-colors"
              onClick={() => window.location.href = '/'}
            >
              ← Back to Home
            </button>
          </div>
        </motion.div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="glass border-b border-white/10 sticky top-0 z-50">
        <div className="max-w-4xl mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <span className="text-2xl">🚂</span>
              <div>
                <h1 className="text-xl font-bold text-white">TrainBlink Chat</h1>
                <p className="text-gray-400" style={{fontSize: '0.875rem'}}>
                  Welcome, {username}!
                </p>
              </div>
            </div>
            <div className="flex items-center space-x-4">
              <div className="flex items-center space-x-2">
                <div className={isConnected ? 'status-online' : 'status-offline'}></div>
                <span className="text-gray-300" style={{fontSize: '0.875rem'}}>
                  {isConnected ? 'Connected' : 'Disconnected'}
                </span>
              </div>
              <button
                onClick={() => window.location.href = '/'}
                className="text-gray-400 hover:text-white transition-colors"
              >
                Home
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Chat Content */}
      <div className="max-w-4xl mx-auto px-4 py-6">
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
          {/* Chat Messages */}
          <div className="lg:col-span-3">
            <div className="glass rounded-lg h-96 flex flex-col">
              {/* Messages Area */}
              <div className="flex-1 p-4 overflow-y-auto">
                {error && (
                  <div className="bg-red-500/20 border border-red-500/30 rounded-lg p-3 mb-4">
                    <p className="text-red-300">{error}</p>
                  </div>
                )}
                
                {messages.length === 0 ? (
                  <div className="text-center text-gray-400 mt-8">
                    <div className="text-3xl mb-2">💭</div>
                    <p>No messages yet. Start the conversation!</p>
                  </div>
                ) : (
                  <div className="space-y-3">
                    {messages.map((msg, index) => (
                      <motion.div
                        key={index}
                        className="bg-white/5 rounded-lg p-3 border border-white/10"
                        initial={{ opacity: 0, y: 10 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ duration: 0.3 }}
                      >
                        <div className="text-white">{msg.message}</div>
                        <div className="text-gray-400 text-xs mt-1">
                          {new Date(msg.timestamp).toLocaleTimeString()}
                        </div>
                      </motion.div>
                    ))}
                  </div>
                )}
              </div>

              {/* Message Input */}
              <form onSubmit={handleSendMessage} className="p-4 border-t border-white/10">
                <div className="flex space-x-3">
                  <input
                    type="text"
                    value={messageInput}
                    onChange={(e) => setMessageInput(e.target.value)}
                    placeholder="Type your message..."
                    className="flex-1 px-4 py-3 rounded-lg bg-white/10 border border-white/20 text-white placeholder-gray-400 focus:outline-none focus:border-blue-400"
                    disabled={!isConnected}
                  />
                  <button
                    type="submit"
                    className="button-primary px-6 py-3"
                    disabled={!isConnected || !messageInput.trim()}
                  >
                    Send 📤
                  </button>
                </div>
                {!isConnected && (
                  <p className="text-red-300 text-sm mt-2">
                    Reconnecting to chat server...
                  </p>
                )}
              </form>
            </div>
          </div>

          {/* Sidebar */}
          <div className="lg:col-span-1">
            <div className="glass rounded-lg p-4">
              <h3 className="font-bold text-white mb-4">Station: Tokyo Central</h3>
              
              <div className="space-y-4">
                <div>
                  <h4 className="text-gray-300 font-medium mb-2">
                    Online Users ({connections?.count || 0})
                  </h4>
                  <div className="space-y-2">
                    <div className="flex items-center space-x-2">
                      <div className="status-online"></div>
                      <span className="text-gray-300">{username}</span>
                      <span className="text-gray-500">(you)</span>
                    </div>
                    {connections?.client_ids?.slice(0, 5).map((id, index) => (
                      <div key={id} className="flex items-center space-x-2">
                        <div className="status-online"></div>
                        <span className="text-gray-400" style={{fontSize: '0.875rem'}}>
                          User {id.slice(-4)}
                        </span>
                      </div>
                    ))}
                    {connections && connections.count > 5 && (
                      <div className="text-gray-500" style={{fontSize: '0.75rem'}}>
                        +{connections.count - 5} more users
                      </div>
                    )}
                  </div>
                </div>

                <div>
                  <h4 className="text-gray-300 font-medium mb-2">Chat Options</h4>
                  <div className="space-y-3">
                    <label className="flex items-center space-x-2 cursor-pointer">
                      <input
                        type="checkbox"
                        checked={usePostApi}
                        onChange={(e) => setUsePostApi(e.target.checked)}
                        className="rounded"
                      />
                      <span className="text-gray-400" style={{fontSize: '0.875rem'}}>
                        Use POST API
                      </span>
                    </label>
                    <div className="text-gray-400 space-y-1" style={{fontSize: '0.75rem'}}>
                      <p>• Anonymous messaging</p>
                      <p>• Real-time updates</p>
                      <p>• Station-based rooms</p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}