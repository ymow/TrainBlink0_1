import { motion } from 'framer-motion';
import { useServerHealth, usePing, useConnections } from '../hooks/useStats';
import Hero from '../components/Hero';
import Features from '../components/Features';
import Stats from '../components/Stats';
import Footer from '../components/Footer';

export default function Home() {
  const { data: health, isLoading: healthLoading } = useServerHealth();
  const { data: ping, isLoading: pingLoading } = usePing();
  const { data: connections, isLoading: connectionsLoading } = useConnections();

  return (
    <div className="min-h-screen">
      {/* Navigation */}
      <nav className="fixed top-0 w-full z-50 glass">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center space-x-2">
              <span className="text-2xl">🚂</span>
              <span className="text-xl font-bold">TrainBlink</span>
            </div>
            <div className="flex items-center space-x-4">
              <button
                className="button-primary"
                onClick={() => window.location.href = '/chat'}
              >
                Try Web Chat
              </button>
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

      {/* Main Content */}
      <main className="pt-16">
        <Hero />
        <Stats 
          health={health} 
          connections={connections?.count} 
          isLoading={healthLoading || connectionsLoading} 
        />
        <Features />
      </main>

      <Footer />
    </div>
  );
}