document.addEventListener('DOMContentLoaded', () => {
    // DOM elements
    const elements = {
        tabs: document.querySelectorAll('.tab'),
        loginForm: document.getElementById('login-form'),
        registerForm: document.getElementById('register-form'),
    };

    // Check if user is already logged in
    function checkAuth() {
        const token = localStorage.getItem('authToken');
        if (token) {
            // User is already logged in, redirect to dashboard
            window.location.href = '/dashboard';
        }
    }

    // Switch between auth tabs
    function switchTab(tabType) {
        elements.tabs.forEach(tab => tab.classList.remove('active'));
        document.querySelector(`.tab[data-tab="${tabType}"]`).classList.add('active');
        
        document.querySelectorAll('.form-container').forEach(form => form.classList.remove('active-form'));
        document.querySelector(`.form-container.${tabType}`).classList.add('active-form');
    }

    // Event listeners
    function attachEventListeners() {
        // Auth tabs
        if (elements.tabs) {
            elements.tabs.forEach(tab => {
                tab.addEventListener('click', () => {
                    const tabType = tab.getAttribute('data-tab');
                    switchTab(tabType);
                });
            });
        }

        // Intercept login form
        if (elements.loginForm) {
            elements.loginForm.addEventListener('submit', async (e) => {
                e.preventDefault();
                const username = document.getElementById('login-username').value;
                const password = document.getElementById('login-password').value;
                
                const loginMessage = document.getElementById('login-message');
                if (loginMessage) {
                    loginMessage.textContent = '';
                    loginMessage.className = 'message';
                }
                
                try {
                    const response = await fetch('/auth/login', {
                        method: 'POST',
                        body: `${username}\n${password}`
                    });
                    
                    if (!response.ok) {
                        throw new Error(await response.text() || 'Login failed');
                    }
                    
                    const token = await response.text();
                    localStorage.setItem('authToken', token);
                    localStorage.setItem('username', username);
                    
                    // Redirect to dashboard
                    window.location.href = '/dashboard';
                    
                } catch (error) {
                    if (loginMessage) {
                        loginMessage.textContent = error.message;
                        loginMessage.className = 'message error';
                    }
                }
            });
        }

        // Intercept register form
        if (elements.registerForm) {
            elements.registerForm.addEventListener('submit', async (e) => {
                e.preventDefault();
                const username = document.getElementById('register-username').value;
                const password = document.getElementById('register-password').value;
                const confirm = document.getElementById('register-confirm').value;
                
                const registerMessage = document.getElementById('register-message');
                
                if (password !== confirm) {
                    if (registerMessage) {
                        registerMessage.textContent = 'Passwords do not match';
                        registerMessage.className = 'message error';
                    }
                    return;
                }
                
                try {
                    const response = await fetch('/auth/register', {
                        method: 'POST',
                        body: `${username}\n${password}`
                    });
                    
                    if (!response.ok) {
                        throw new Error(await response.text() || 'Registration failed');
                    }
                    
                    if (registerMessage) {
                        registerMessage.textContent = 'Registration successful! You can now log in.';
                        registerMessage.className = 'message success';
                    }
                    
                    setTimeout(() => {
                        switchTab('login');
                    }, 2000);
                    
                } catch (error) {
                    if (registerMessage) {
                        registerMessage.textContent = error.message;
                        registerMessage.className = 'message error';
                    }
                }
            });
        }
    }

    // Initialize the page
    attachEventListeners();
    checkAuth();
});
