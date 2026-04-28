// App.jsx
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import { RegistrationPage } from './pages/RegistrationPage'
import { VerifyPage } from './pages/VerifyPage'

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/register" element={<RegistrationPage/>}/>
        <Route path="/verify" element={<VerifyPage />} />
      </Routes>
    </Router>
  )
}

export default App