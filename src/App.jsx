import "./App.css";
import Navbar from "./components/Navbar";
import Section1 from "./sections/Section1";
import { ToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';

function App() {

  return (
    <div style={{height:"100vh", zIndex:100}}>
      <Navbar/>
      <Section1/>
      <ToastContainer />
    </div>
  );
}

export default App;
