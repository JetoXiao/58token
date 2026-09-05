/** @type {import('tailwindcss').Config} */
const primaryPalette = {
  50: '#eff8ff',
  100: '#dbeeff',
  200: '#b9dcff',
  300: '#8ac7ff',
  400: '#58adff',
  500: '#2f8bff',
  600: '#1f6fe6',
  700: '#1e5cc0',
  800: '#1f4b9a',
  900: '#1d3f7b',
  950: '#11244d'
}

const accentPalette = {
  50: '#f6f3ff',
  100: '#efe9ff',
  200: '#ddd4ff',
  300: '#c4b5fd',
  400: '#a78bfa',
  500: '#8b5cf6',
  600: '#7c3aed',
  700: '#6d28d9',
  800: '#5b21b6',
  900: '#4c1d95',
  950: '#2e1065'
}

const cyanPalette = {
  50: '#effdff',
  100: '#daf8ff',
  200: '#b8efff',
  300: '#87e2ff',
  400: '#59cff8',
  500: '#2fb3e9',
  600: '#2491c8',
  700: '#2273a0',
  800: '#215c81',
  900: '#1e4b66',
  950: '#102d40'
}

const warmPalette = {
  50: '#fff7f2',
  100: '#ffeadd',
  200: '#ffd3be',
  300: '#ffb88e',
  400: '#ff9d68',
  500: '#ff8149',
  600: '#e96730',
  700: '#c64f26',
  800: '#9f4020',
  900: '#80341d',
  950: '#46190e'
}

export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        // 品牌色族统一映射：蓝青紫 + 暖橙光晕
        primary: primaryPalette,
        accent: accentPalette,
        cyan: cyanPalette,
        blue: primaryPalette,
        sky: cyanPalette,
        teal: cyanPalette,
        emerald: cyanPalette,
        green: cyanPalette,
        lime: cyanPalette,
        purple: accentPalette,
        violet: accentPalette,
        fuchsia: accentPalette,
        pink: accentPalette,
        amber: warmPalette,
        yellow: warmPalette,
        orange: warmPalette,
        rose: warmPalette,
        // 深色模式背景
        dark: {
          50: '#f8fafc',
          100: '#f1f5f9',
          200: '#e2e8f0',
          300: '#cbd5e1',
          400: '#94a3b8',
          500: '#64748b',
          600: '#475569',
          700: '#334155',
          800: '#1e293b',
          900: '#0f172a',
          950: '#020617'
        }
      },
      fontFamily: {
        sans: [
          'Inter',
          'system-ui',
          '-apple-system',
          'BlinkMacSystemFont',
          'Segoe UI',
          'Roboto',
          'Helvetica Neue',
          'Arial',
          'PingFang SC',
          'Hiragino Sans GB',
          'Microsoft YaHei',
          'sans-serif'
        ],
        mono: ['ui-monospace', 'SFMono-Regular', 'Menlo', 'Monaco', 'Consolas', 'monospace']
      },
      boxShadow: {
        glass: '0 24px 80px rgba(0, 0, 0, 0.22)',
        'glass-sm': '0 10px 32px rgba(0, 0, 0, 0.16)',
        glow: '0 0 28px rgba(47, 139, 255, 0.24)',
        'glow-lg': '0 0 64px rgba(139, 92, 246, 0.24)',
        card: '0 18px 60px rgba(0, 0, 0, 0.22)',
        'card-hover': '0 28px 90px rgba(0, 0, 0, 0.32)',
        'inner-glow': 'inset 0 1px 0 rgba(255, 255, 255, 0.1)'
      },
      backgroundImage: {
        'gradient-radial': 'radial-gradient(var(--tw-gradient-stops))',
        'gradient-primary': 'linear-gradient(135deg, #2f8bff 0%, #22b8f0 100%)',
        'gradient-dark': 'linear-gradient(135deg, #1e293b 0%, #0f172a 100%)',
        'gradient-glass':
          'linear-gradient(135deg, rgba(255,255,255,0.1) 0%, rgba(255,255,255,0.05) 100%)',
        'mesh-gradient':
          'radial-gradient(at 14% 10%, rgba(214, 231, 255, 0.88) 0px, transparent 30%), radial-gradient(at 50% 4%, rgba(228, 220, 255, 0.72) 0px, transparent 28%), radial-gradient(at 86% 12%, rgba(255, 216, 196, 0.72) 0px, transparent 28%), radial-gradient(at 12% 82%, rgba(215, 242, 255, 0.66) 0px, transparent 28%)'
      },
      animation: {
        'fade-in': 'fadeIn 0.3s ease-out',
        'slide-up': 'slideUp 0.3s ease-out',
        'slide-down': 'slideDown 0.3s ease-out',
        'slide-in-right': 'slideInRight 0.3s ease-out',
        'scale-in': 'scaleIn 0.2s ease-out',
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        shimmer: 'shimmer 2s linear infinite',
        glow: 'glow 2s ease-in-out infinite alternate'
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' }
        },
        slideUp: {
          '0%': { opacity: '0', transform: 'translateY(10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' }
        },
        slideDown: {
          '0%': { opacity: '0', transform: 'translateY(-10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' }
        },
        slideInRight: {
          '0%': { opacity: '0', transform: 'translateX(20px)' },
          '100%': { opacity: '1', transform: 'translateX(0)' }
        },
        scaleIn: {
          '0%': { opacity: '0', transform: 'scale(0.95)' },
          '100%': { opacity: '1', transform: 'scale(1)' }
        },
        shimmer: {
          '0%': { backgroundPosition: '-200% 0' },
          '100%': { backgroundPosition: '200% 0' }
        },
        glow: {
          '0%': { boxShadow: '0 0 20px rgba(47, 139, 255, 0.25)' },
          '100%': { boxShadow: '0 0 30px rgba(34, 184, 240, 0.4)' }
        }
      },
      backdropBlur: {
        xs: '2px'
      },
      borderRadius: {
        '4xl': '2rem'
      }
    }
  },
  plugins: []
}
