import { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { useWebSocket } from '../hooks/useWebSocket';
import { useConnections } from '../hooks/useStats';
import { TrainButton, TrainInput, TrainDialog, Card, Badge } from './ui/BaseUIComponents';

interface ChatInterfaceProps {
  username: string;
  onBack: () => void;
}

export function ChatInterface({ username, onBack }: ChatInterfaceProps) {
  const [messageInput, setMessageInput] = useState('');
  const [showSettings, setShowSettings] = useState(false);
  const [showUserList, setShowUserList] = useState(false);
  
  const { data: connections } = useConnections();
  const { messages, isConnected, error, sendMessage } = useWebSocket('ws://localhost:8080/ws');

  const handleSendMessage = (e: React.FormEvent) => {
    e.preventDefault();
    if (messageInput.trim() && isConnected) {
      sendMessage(`[${username}]: ${messageInput.trim()}`);
      setMessageInput('');
    }
  };

  return (
    <div className="min-h-screen bg-background pt-20">
      {/* Chat Header */}
      <div className="border-b border-white/10">
        <div className="max-w-4xl mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <TrainButton variant="ghost" onClick={onBack}>
                ← Back
              </TrainButton>
              <div>
                <h1 className="text-xl font-bold text-white">🚉 Tokyo Central Station</h1>
                <p className="text-gray-400" style={{fontSize: '0.875rem'}}>
                  Welcome, {username}! • {connections?.count || 0} users online
                </p>
              </div>
            </div>
            <div className="flex items-center space-x-3">
              <Badge variant={isConnected ? 'success' : 'error'}>
                {isConnected ? '✅ Connected' : '❌ Offline'}
              </Badge>
              <TrainButton
                variant="secondary"
                size="sm"
                onClick={() => setShowUserList(true)}
              >
                👥 Users ({connections?.count || 0})
              </TrainButton>
              <TrainButton
                variant="ghost"
                size="sm"
                onClick={() => setShowSettings(true)}
              >
                ⚙️
              </TrainButton>
            </div>
          </div>
        </div>
      </div>

      <div className="max-w-4xl mx-auto px-4 py-6">
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
          {/* Main Chat Area */}
          <div className="lg:col-span-3">
            <Card className="h-96 flex flex-col">
              {/* Messages */}
              <div className="flex-1 p-4 overflow-y-auto">
                {error && (
                  <Card className="p-3 mb-4 bg-red-500/20 border-red-500/30">
                    <Badge variant="error">{error}</Badge>
                  </Card>
                )}
                
                <AnimatePresence>
                  {messages.length === 0 ? (
                    <motion.div
                      className="text-center text-gray-400 mt-8"
                      initial={{ opacity: 0 }}
                      animate={{ opacity: 1 }}
                    >
                      <div className="text-3xl mb-2">💭</div>
                      <p>No messages yet. Start the conversation!</p>
                      <p className="text-sm mt-2">Station chat is anonymous and ephemeral</p>
                    </motion.div>
                  ) : (
                    <div className="space-y-3">
                      {messages.map((msg, index) => (
                        <motion.div
                          key={index}
                          initial={{ opacity: 0, y: 10, scale: 0.95 }}
                          animate={{ opacity: 1, y: 0, scale: 1 }}
                          exit={{ opacity: 0, scale: 0.95 }}
                          transition={{ duration: 0.3 }}
                        >
                          <Card className="p-3 bg-white/5">
                            <div className="text-white">{msg.message}</div>
                            <div className="text-gray-400 text-xs mt-1 flex items-center space-x-2">
                              <span>{new Date(msg.timestamp).toLocaleTimeString()}</span>
                              {msg.type && <Badge variant="info">{msg.type}</Badge>}
                            </div>
                          </Card>
                        </motion.div>
                      ))}
                    </div>
                  )}
                </AnimatePresence>
              </div>

              {/* Message Input */}
              <form onSubmit={handleSendMessage} className="p-4 border-t border-white/10">
                <div className="flex space-x-3">
                  <TrainInput
                    value={messageInput}
                    onChange={(e) => setMessageInput(e.target.value)}
                    placeholder={isConnected ? "Type your message..." : "Connecting..."}
                    disabled={!isConnected}
                    className="flex-1"
                  />
                  <TrainButton
                    type="submit"
                    variant="primary"
                    disabled={!isConnected || !messageInput.trim()}
                  >
                    Send 📤
                  </TrainButton>
                </div>
                {!isConnected && (
                  <motion.p 
                    className="text-red-300 text-sm mt-2"
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                  >
                    🔄 Reconnecting to chat server...
                  </motion.p>
                )}
              </form>
            </Card>
          </div>

          {/* Sidebar */}
          <div className="lg:col-span-1 space-y-4">
            {/* Station Info */}
            <Card className="p-4">
              <h3 className="font-bold text-white mb-3">🚉 Station Info</h3>
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-gray-300">Location:</span>
                  <span className="text-white">Tokyo Central</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-300">Line:</span>
                  <span className="text-white">JR Yamanote</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-300">Platform:</span>
                  <span className="text-white">5-6</span>
                </div>
              </div>
            </Card>

            {/* Quick Actions */}
            <Card className="p-4">
              <h3 className="font-bold text-white mb-3">🚀 Quick Actions</h3>
              <div className="space-y-2">
                <TrainButton
                  variant="secondary"
                  size="sm"
                  className="w-full"
                  onClick={() => setShowUserList(true)}
                >
                  👥 View All Users
                </TrainButton>
                <TrainButton
                  variant="ghost"
                  size="sm"
                  className="w-full"
                  onClick={() => {
                    const helpMsg = "Welcome to TrainBlink! This is an anonymous chat for train station travelers.";
                    sendMessage(`[System]: ${helpMsg}`);
                  }}
                  disabled={!isConnected}
                >
                  ❓ Help
                </TrainButton>
              </div>
            </Card>

            {/* Chat Rules */}
            <Card className="p-4">
              <h3 className="font-bold text-white mb-3">📋 Chat Rules</h3>
              <div className="text-gray-400 space-y-1" style={{fontSize: '0.75rem'}}>
                <p>• Be respectful to fellow travelers</p>
                <p>• No personal information sharing</p>
                <p>• Anonymous chat only</p>
                <p>• Station-based conversations</p>
                <p>• Messages auto-delete after 2 hours</p>
              </div>
            </Card>
          </div>
        </div>
      </div>

      {/* User List Dialog */}
      <TrainDialog 
        open={showUserList} 
        onOpenChange={setShowUserList}
        title="👥 Online Users"
      >
        <div className="space-y-3">
          <div className="flex justify-between items-center">
            <span className="text-gray-300">Total Users:</span>
            <Badge variant="success">{connections?.count || 0}</Badge>
          </div>
          
          <div className="max-h-48 overflow-y-auto">
            <div className="space-y-2">
              <div className="flex items-center space-x-2 p-2 glass rounded">
                <div className="status-online"></div>
                <span className="text-white">{username}</span>
                <Badge variant="info">You</Badge>
              </div>
              
              {connections?.client_ids?.slice(0, 10).map((id, index) => (
                <div key={id} className="flex items-center space-x-2 p-2 glass rounded">
                  <div className="status-online"></div>
                  <span className="text-gray-300">Traveler {id.slice(-4)}</span>
                </div>
              ))}
              
              {connections && connections.count > 10 && (
                <p className="text-center text-gray-500 text-sm">
                  +{connections.count - 10} more travelers...
                </p>
              )}
            </div>
          </div>
          
          <TrainButton 
            variant="secondary" 
            className="w-full"
            onClick={() => setShowUserList(false)}
          >
            Close
          </TrainButton>
        </div>
      </TrainDialog>

      {/* Settings Dialog */}
      <TrainDialog 
        open={showSettings} 
        onOpenChange={setShowSettings}
        title="⚙️ Chat Settings"
      >
        <div className="space-y-4">
          <div>
            <h4 className="text-white font-medium mb-2">Notifications</h4>
            <label className="flex items-center space-x-2">
              <input type="checkbox" defaultChecked className="rounded" />
              <span className="text-gray-300">Sound alerts</span>
            </label>
          </div>
          
          <div>
            <h4 className="text-white font-medium mb-2">Display</h4>
            <label className="flex items-center space-x-2">
              <input type="checkbox" defaultChecked className="rounded" />
              <span className="text-gray-300">Show timestamps</span>
            </label>
          </div>
          
          <div className="flex space-x-2">
            <TrainButton 
              variant="secondary" 
              className="flex-1"
              onClick={() => setShowSettings(false)}
            >
              Cancel
            </TrainButton>
            <TrainButton 
              variant="primary" 
              className="flex-1"
              onClick={() => setShowSettings(false)}
            >
              Save
            </TrainButton>
          </div>
        </div>
      </TrainDialog>
    </div>
  );
}