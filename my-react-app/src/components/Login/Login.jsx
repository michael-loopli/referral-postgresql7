import { useState } from 'react';
import axios from 'axios';
import user_icon from '/src/assets/person.png';
import password_icon from '/src/assets/password.png';
import './Login.css';

const Login = () => {
    const [credentials, setCredentials] = useState({
        email: '',
        password: ''
    });

    const handleChange = (e) => {
        setCredentials({
            ...credentials,
            [e.target.name]: e.target.value
        });
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        try {
            const response = await axios.post('http://localhost:8080/login', credentials, { withCredentials: true });
            if (response.status === 200) {
                const { role } = response.data; // Assume the API returns the user role
                localStorage.setItem('userRole', role); // Store the role in localStorage
                alert('Login successful');
                window.location.href = "/referrals";
            }
        } catch (error) {
            console.error('There was an error logging in!', error);
            alert('Login failed: ' + error.response?.data || error.message);
        }
    };

    return (
        <div className='container'>
            <div className='header'>
                <div className='text'>Login</div>
                <div className='underline'></div>
            </div>
            <form onSubmit={handleSubmit}>
                <div className='inputs'>
                    <div>
                        <div className='input'>
                            <img src={user_icon} alt='user' />
                            <input name="email" placeholder="Email" onChange={handleChange} />
                        </div>
                        <br />
                    </div>
                    <div className='input'>
                        <img src={password_icon} alt='pswd' />
                        <input name="password" type="password" placeholder="Password" onChange={handleChange} />
                    </div>
                    <div className='submit-container'>
                        <button className='submit' type="submit">Login</button>
                    </div>
                </div>
            </form>
        </div>
    );
};

export default Login;
