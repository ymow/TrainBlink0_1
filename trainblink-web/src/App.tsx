import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { useState } from 'react';
import { motion } from 'framer-motion';
import { useServerHealth, useConnections, usePing } from './hooks/useStats';
import { TrainButton, TrainInput, Card, Badge } from './components/ui/BaseUIComponents';
import { ChatInterface } from './components/ChatInterface';
import './index.css';

const queryClient = new QueryClient();

// Navigation Component
function Navigation() {
  const { data: ping, isLoading: pingLoading } = usePing();
  
  return (
    <nav className="fixed top-0 w-full z-50 glass">
      <div className="max-w-7xl mx-auto px-4 py-4">
        <div className="flex justify-between items-center">
          <div className="flex items-center space-x-2">
            <span className="text-2xl">🚂</span>
            <span className="text-xl font-bold text-white">TrainBlink</span>
          </div>
          <div className="flex items-center space-x-4">
            <TrainButton 
              variant="primary"
              onClick={() => window.location.href = '/chat'}
            >
              Try Chat
            </TrainButton>
            <div className="flex items-center space-x-2">
              <div className={ping ? 'status-online' : 'status-offline'}></div>
              <span className="text-gray-300" style={{fontSize: '0.875rem'}}>
                {pingLoading ? 'Connecting...' : ping ? 'Online' : 'Offline'}
              </span>
            </div>
          </div>
        </div>
      </div>
    </nav>
  );
}

// Home Page Component
function Home() {
  const { data: health, isLoading: healthLoading } = useServerHealth();
  const { data: connections } = useConnections();

  const formatUptime = (seconds: number) => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return `${hours}h ${minutes}m`;
  };

  const stats = [
    {
      title: 'Train Stations',
      value: '253',
      icon: '🚉',
      subtitle: 'Japan & Taiwan'
    },
    {
      title: 'Active Users',
      value: healthLoading ? '...' : (connections?.count || 0),
      icon: '👥',
      subtitle: 'Currently online'
    },
    {
      title: 'Total Requests',
      value: healthLoading ? '...' : (health?.total_requests || 0),
      icon: '📊',
      subtitle: 'Server requests'
    },
    {
      title: 'Server Uptime',
      value: healthLoading ? '...' : (health?.uptime ? formatUptime(health.uptime) : '0h 0m'),
      icon: '⏰',
      subtitle: 'Continuous operation'
    }
  ];

  return (
    <div className="min-h-screen pt-20">
      {/* Hero Section */}
      <section className="relative min-h-screen flex items-center justify-center overflow-hidden">
        {/* Background Animation */}
        <div className="absolute inset-0">
          <div className="absolute inset-0 bg-gradient-to-br from-primary-500/20 via-background to-accent-500/10" />
          <motion.div
            className="absolute top-1/4 left-1/4 w-64 h-64 bg-primary-500/10 rounded-full blur-3xl"
            animate={{
              x: [0, 100, 0],
              y: [0, -50, 0],
            }}
            transition={{
              duration: 20,
              repeat: Infinity,
              ease: "linear"
            }}
          />
        </div>

        {/* Content */}
        <div className="relative z-10 max-w-6xl mx-auto px-4 text-center">
          <motion.div
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8 }}
          >
            {/* Train Icon */}
            <motion.div
              className="flex justify-center mb-8"
              initial={{ opacity: 0, scale: 0.8 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ duration: 1, delay: 0.2 }}
            >
              <motion.span
                className="text-8xl md:text-9xl"
                animate={{ rotateY: [0, 10, -10, 0] }}
                transition={{ duration: 4, repeat: Infinity, ease: "easeInOut" }}
              >
                🚂
              </motion.span>
            </motion.div>

            <motion.h1
              className="text-5xl md:text-7xl font-bold mb-6 bg-gradient-to-r from-white to-primary-300 text-white"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.8, delay: 0.4 }}
            >
              TrainBlink
            </motion.h1>

            <motion.p
              className="text-xl md:text-2xl text-gray-300 mb-8 max-w-3xl mx-auto"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.8, delay: 0.6 }}
            >
              Connect with fellow travelers on trains and at stations.
              <br />
              <span className="text-accent-500 font-semibold">
                Anonymous • Secure • Real-time
              </span>
            </motion.p>

            <motion.div
              className="flex flex-wrap justify-center gap-4 mb-8"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.8, delay: 0.8 }}
            >
              {['🔒 Anonymous Chat', '📍 Location-Based', '⚡ Real-time', '🚄 Cross-Platform'].map((feature, index) => (
                <motion.div
                  key={feature}
                  whileHover={{ scale: 1.05 }}
                >
                  <Badge variant="info">
                    {feature}
                  </Badge>
                </motion.div>
              ))}
            </motion.div>

            <motion.div
              className="flex flex-col sm:flex-row gap-4 justify-center items-center"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.8, delay: 1.0 }}
            >
              <TrainButton
                variant="primary"
                size="lg"
                onClick={() => window.location.href = '/chat'}
              >
                🚀 Try Web Chat Now
              </TrainButton>
            </motion.div>
          </motion.div>
        </div>
      </section>

      {/* Stats Section */}
      <section className="py-20 px-4">
        <div className="max-w-7xl mx-auto">
          <div className="text-center mb-16">
            <h2 className="text-3xl md:text-4xl font-bold text-white mb-4">
              Live Network Statistics
            </h2>
            <p className="text-gray-300 text-lg">
              Real-time data from our TrainBlink network
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            {stats.map((stat, index) => (
              <motion.div
                key={stat.title}
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.6, delay: index * 0.1 }}
                viewport={{ once: true }}
              >
                <Card hover className="p-6 text-center">
                  <div className="text-4xl mb-4">{stat.icon}</div>
                  <div className="text-3xl font-bold text-white mb-2">
                    {typeof stat.value === 'number' ? stat.value.toLocaleString() : stat.value}
                  </div>
                  <div className="text-gray-300 font-medium">{stat.title}</div>
                  <div className="text-sm text-gray-400 mt-1">{stat.subtitle}</div>
                </Card>
              </motion.div>
            ))}
          </div>

          <motion.div
            className="mt-12 text-center"
            initial={{ opacity: 0 }}
            whileInView={{ opacity: 1 }}
            transition={{ duration: 0.8, delay: 0.5 }}
            viewport={{ once: true }}
          >
            <div className="glass inline-flex items-center space-x-3 px-6 py-3 rounded-full">
              <motion.div
                className="w-3 h-3 bg-green-400 rounded-full"
                animate={{ opacity: [1, 0.5, 1] }}
                transition={{ duration: 2, repeat: Infinity }}
              />
              <span className="text-white font-medium">
                Server is online and ready
              </span>
            </div>
          </motion.div>
        </div>
      </section>
    </div>
  );
}

// Chat Page Component  
function Chat() {
  const [username, setUsername] = useState('');
  const [isUsernameSet, setIsUsernameSet] = useState(false);

  const handleSetUsername = (e: React.FormEvent) => {
    e.preventDefault();
    if (username.trim()) {
      setIsUsernameSet(true);
    }
  };

  const handleBack = () => {
    setIsUsernameSet(false);
    setUsername('');
  };

  if (!isUsernameSet) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center px-4">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8 }}
        >
          <Card className="p-8 w-full max-w-md">
            <div className="text-center mb-6">
              <div className="text-4xl mb-4">🚂</div>
              <h1 className="text-2xl font-bold text-white mb-2">Join TrainBlink Chat</h1>
              <p className="text-gray-300">Enter a username to start chatting</p>
            </div>
            
            <form onSubmit={handleSetUsername}>
              <TrainInput
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="Your username..."
                maxLength={20}
                required
                className="mb-4"
              />
              <TrainButton
                type="submit"
                variant="primary"
                size="lg"
                disabled={!username.trim()}
                className="w-full"
              >
                Start Chatting 💬
              </TrainButton>
            </form>
            
            <div className="mt-4 text-center">
              <TrainButton
                variant="ghost"
                onClick={() => window.location.href = '/'}
              >
                ← Back to Home
              </TrainButton>
            </div>
          </Card>
        </motion.div>
      </div>
    );
  }

  return <ChatInterface username={username} onBack={handleBack} />;
}

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <Router>
        <div className="min-h-screen bg-background">
          <Navigation />
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/chat" element={<Chat />} />
          </Routes>
        </div>
      </Router>
    </QueryClientProvider>
  );
}

export default App;