import { ThemeProvider } from './context/ThemeContext';
import { Dashboard } from './pages/Dashboard';
import './styles/theme.css';

function App() {
  return (
    <ThemeProvider>
      <Dashboard />
    </ThemeProvider>
  );
}

export default App;
