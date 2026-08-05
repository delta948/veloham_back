export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        acid: '#b6ff00',
        danger: '#ff2b2b',
        asphalt: '#0b0d0d',
        steel: '#151818'
      },
      boxShadow: {
        street: '0 18px 55px rgba(182,255,0,.12)',
        danger: '0 18px 55px rgba(255,43,43,.12)'
      }
    }
  },
  plugins: []
};
