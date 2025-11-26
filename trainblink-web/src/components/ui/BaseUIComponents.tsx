// Base UI Component Wrappers
import React from 'react';

// Button Component with TrainBlink styling
interface ButtonProps {
  children: React.ReactNode;
  onClick?: () => void;
  variant?: 'primary' | 'secondary' | 'ghost';
  size?: 'sm' | 'md' | 'lg';
  disabled?: boolean;
  type?: 'button' | 'submit' | 'reset';
  className?: string;
}

export function TrainButton({ 
  children, 
  onClick, 
  variant = 'primary', 
  size = 'md', 
  disabled = false,
  type = 'button',
  className = ''
}: ButtonProps) {
  const baseClasses = 'transition-all duration-300 font-medium rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500';
  
  const variantClasses = {
    primary: 'bg-primary-500 hover:bg-primary-600 text-white shadow-lg hover:shadow-xl transform hover:scale-105',
    secondary: 'glass border border-white/20 text-white hover:bg-white/20',
    ghost: 'text-gray-300 hover:text-white hover:bg-white/10'
  };

  const sizeClasses = {
    sm: 'px-3 py-2 text-sm',
    md: 'px-4 py-3 text-base',
    lg: 'px-6 py-4 text-lg'
  };

  return (
    <button
      onClick={onClick}
      disabled={disabled}
      type={type}
      className={`${baseClasses} ${variantClasses[variant]} ${sizeClasses[size]} ${className}`}
    >
      {children}
    </button>
  );
}

// Input Component with TrainBlink styling
interface InputProps {
  value: string;
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  placeholder?: string;
  type?: string;
  disabled?: boolean;
  required?: boolean;
  maxLength?: number;
  className?: string;
}

export function TrainInput({ 
  value, 
  onChange, 
  placeholder = '', 
  type = 'text',
  disabled = false,
  required = false,
  maxLength,
  className = ''
}: InputProps) {
  const baseClasses = 'w-full px-4 py-3 glass border border-white/20 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 transition-all duration-300';

  return (
    <input
      value={value}
      onChange={onChange}
      placeholder={placeholder}
      type={type}
      disabled={disabled}
      required={required}
      maxLength={maxLength}
      className={`${baseClasses} ${className}`}
    />
  );
}

// Dialog Component for modals
interface DialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  children: React.ReactNode;
  title?: string;
}

export function TrainDialog({ open, onOpenChange, children, title }: DialogProps) {
  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div 
        className="fixed inset-0 z-40 bg-black/50 backdrop-blur-sm" 
        onClick={() => onOpenChange(false)}
      />
      <div className="relative z-50 w-full max-w-md mx-4">
        <div className="glass rounded-xl p-6 border border-white/20 shadow-xl">
          {title && (
            <h2 className="text-xl font-bold text-white mb-4">
              {title}
            </h2>
          )}
          {children}
        </div>
      </div>
    </div>
  );
}

// Card Component
interface CardProps {
  children: React.ReactNode;
  className?: string;
  hover?: boolean;
}

export function Card({ children, className = '', hover = false }: CardProps) {
  const baseClasses = 'glass rounded-lg border border-white/20 transition-all duration-300';
  const hoverClasses = hover ? 'hover:scale-105 hover:shadow-xl' : '';
  
  return (
    <div className={`${baseClasses} ${hoverClasses} ${className}`}>
      {children}
    </div>
  );
}

// Badge Component
interface BadgeProps {
  children: React.ReactNode;
  variant?: 'success' | 'warning' | 'error' | 'info';
}

export function Badge({ children, variant = 'info' }: BadgeProps) {
  const variantClasses = {
    success: 'bg-green-500/20 text-green-300 border-green-500/30',
    warning: 'bg-yellow-500/20 text-yellow-300 border-yellow-500/30',
    error: 'bg-red-500/20 text-red-300 border-red-500/30',
    info: 'bg-blue-500/20 text-blue-300 border-blue-500/30'
  };

  return (
    <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium border ${variantClasses[variant]}`}>
      {children}
    </span>
  );
}