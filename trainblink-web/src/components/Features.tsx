import { motion } from 'framer-motion';

interface FeatureProps {
  title: string;
  description: string;
  icon: string;
  delay: number;
}

function FeatureCard({ title, description, icon, delay }: FeatureProps) {
  return (
    <motion.div
      className="glass p-8 rounded-2xl hover:bg-white/15 transition-all duration-300"
      initial={{ opacity: 0, y: 30 }}
      whileInView={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.6, delay }}
      whileHover={{ scale: 1.02, y: -5 }}
      viewport={{ once: true }}
    >
      <div className="text-5xl mb-6 text-center">{icon}</div>
      <h3 className="text-xl font-bold text-white mb-4 text-center">{title}</h3>
      <p className="text-gray-300 leading-relaxed text-center">{description}</p>
    </motion.div>
  );
}

export default function Features() {
  const features = [
    {
      title: 'Anonymous Chat',
      description: 'Connect without revealing your identity. Every conversation is private and secure.',
      icon: '🔒'
    },
    {
      title: 'Location-Based Rooms',
      description: 'Join chat rooms based on your current train station or route.',
      icon: '📍'
    },
    {
      title: 'Real-time Messaging',
      description: 'Instant communication with fellow travelers using WebSocket technology.',
      icon: '⚡'
    },
    {
      title: 'Cross-Platform',
      description: 'Available on web, iOS, and Android. Chat from any device, anywhere.',
      icon: '📱'
    },
    {
      title: 'BLE Discovery',
      description: 'Discover nearby travelers using Bluetooth Low Energy when offline.',
      icon: '📶'
    },
    {
      title: 'Matrix Integration',
      description: 'Built on Matrix protocol for decentralized, secure communications.',
      icon: '🔗'
    }
  ];

  return (
    <section className="py-20 px-4 sm:px-6 lg:px-8">
      <div className="max-w-7xl mx-auto">
        {/* Section Header */}
        <motion.div
          className="text-center mb-16"
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8 }}
          viewport={{ once: true }}
        >
          <h2 className="text-3xl md:text-4xl lg:text-5xl font-bold text-white mb-6">
            Connect. Chat. Travel.
          </h2>
          <p className="text-gray-300 text-lg md:text-xl max-w-3xl mx-auto leading-relaxed">
            TrainBlink revolutionizes how travelers connect during their journeys.
            Experience seamless communication with privacy and security at its core.
          </p>
        </motion.div>

        {/* Features Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
          {features.map((feature, index) => (
            <FeatureCard
              key={feature.title}
              {...feature}
              delay={index * 0.1}
            />
          ))}
        </div>

        {/* How It Works Section */}
        <motion.div
          className="mt-20"
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          transition={{ duration: 0.8 }}
          viewport={{ once: true }}
        >
          <h3 className="text-2xl md:text-3xl font-bold text-white text-center mb-12">
            How TrainBlink Works
          </h3>
          
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            {[
              {
                step: '01',
                title: 'Enter Station',
                description: 'Walk into any supported train station and TrainBlink detects your location',
                icon: '🚉'
              },
              {
                step: '02',
                title: 'Join Room',
                description: 'Automatically join the station chat room or browse available conversations',
                icon: '💬'
              },
              {
                step: '03',
                title: 'Start Chatting',
                description: 'Connect with other travelers, share tips, or just have a friendly conversation',
                icon: '✨'
              }
            ].map((item, index) => (
              <motion.div
                key={item.step}
                className="text-center"
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.6, delay: index * 0.2 }}
                viewport={{ once: true }}
              >
                <div className="relative mb-6">
                  <div className="w-16 h-16 bg-primary-500/20 rounded-full flex items-center justify-center mx-auto mb-4">
                    <span className="text-2xl">{item.icon}</span>
                  </div>
                  <div className="absolute -top-2 -right-2 w-8 h-8 bg-accent-500 text-background rounded-full flex items-center justify-center text-sm font-bold">
                    {item.step}
                  </div>
                </div>
                <h4 className="text-lg font-semibold text-white mb-3">{item.title}</h4>
                <p className="text-gray-300">{item.description}</p>
              </motion.div>
            ))}
          </div>
        </motion.div>
      </div>
    </section>
  );
}