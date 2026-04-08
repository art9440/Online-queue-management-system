// App.jsx
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import { RegistrationPage } from './pages/RegistrationPage'

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/register" element={<RegistrationPage/>}/>
      </Routes>
    </Router>
  )
}

export default App