// App.jsx
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import { RegistrationPage } from './pages/RegistrationPage'
import { PublicBookingPage } from './pages/PublicBookingPage'
import { VerifyPage } from './pages/VerifyPage'
import { LoginPage } from './pages/LoginPage'

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/register" element={<RegistrationPage/>}/>
        <Route path="/verify" element={<VerifyPage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/public/:registrationSlug" element={<PublicBookingPage />} />
      </Routes>
    </Router>
  )
}

export default App
