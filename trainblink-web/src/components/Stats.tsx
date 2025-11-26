import { motion } from 'framer-motion';
import { type HealthStatus } from '../services/api';

interface StatCardProps {
  title: string;
  value: number | string;
  icon: string;
  delay: number;
  subtitle?: string;
}

function StatCard({ title, value, icon, delay, subtitle }: StatCardProps) {
  return (
    <motion.div
      className="glass p-6 rounded-2xl text-center"
      initial={{ opacity: 0, y: 20 }}
      whileInView={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.6, delay }}
      whileHover={{ scale: 1.05 }}
      viewport={{ once: true }}
    >
      <div className="text-4xl mb-4">{icon}</div>
      <motion.div
        className="text-3xl md:text-4xl font-bold text-white mb-2"
        initial={{ scale: 0 }}
        whileInView={{ scale: 1 }}
        transition={{ duration: 0.5, delay: delay + 0.2 }}
        viewport={{ once: true }}
      >
        {typeof value === 'number' ? value.toLocaleString() : value}
      </motion.div>
      <div className="text-gray-300 font-medium">{title}</div>
      {subtitle && (
        <div className="text-sm text-gray-400 mt-1">{subtitle}</div>
      )}
    </motion.div>
  );
}

interface StatsProps {
  health?: HealthStatus;
  connections?: number;
  isLoading: boolean;
}

export default function Stats({ health, connections, isLoading }: StatsProps) {
  const formatUptime = (seconds: number) => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return `${hours}h ${minutes}m`;
  };

  const displayStats = [
    {
      title: 'Train Stations',
      value: 253,
      icon: '🚉',
      subtitle: 'Japan & Taiwan'
    },
    {
      title: 'Active Connections',
      value: isLoading ? '...' : (connections || health?.connections || 0),
      icon: '👥',
      subtitle: 'Currently online'
    },
    {
      title: 'Total Requests',
      value: isLoading ? '...' : (health?.total_requests || 0),
      icon: '📊',
      subtitle: 'Server requests'
    },
    {
      title: 'Server Uptime',
      value: isLoading ? '...' : (health?.uptime ? formatUptime(health.uptime) : '0h 0m'),
      icon: '⏰',
      subtitle: 'Continuous operation'
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
          <h2 className="text-3xl md:text-4xl font-bold text-white mb-4">
            Live Network Statistics
          </h2>
          <p className="text-gray-300 text-lg max-w-2xl mx-auto">
            Real-time data from our growing community of train travelers
          </p>
        </motion.div>

        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {displayStats.map((stat, index) => (
            <StatCard
              key={stat.title}
              {...stat}
              delay={index * 0.1}
            />
          ))}
        </div>

        {/* Server Status */}
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
              {isLoading ? 'Connecting to server...' : 'Server is online and ready'}
            </span>
          </div>
        </motion.div>
      </div>
    </section>
  );
}