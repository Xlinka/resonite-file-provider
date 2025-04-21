document.addEventListener('DOMContentLoaded', () => {
    // DOM elements
    const elements = {
        authSection: document.getElementById('auth-section'),
        dashboardSection: document.getElementById('dashboard-section'),
        tabs: document.querySelectorAll('.tab'),
        loginForm: document.getElementById('login-form'),
        registerForm: document.getElementById('register-form'),
        uploadBtn: document.getElementById('upload-btn'),
        uploadModal: document.getElementById('upload-modal'),
        newFolderBtn: document.getElementById('new-folder-btn'),
        newFolderModal: document.getElementById('new-folder-modal'),
        closeButtons: document.querySelectorAll('.close'),
        currentFolderId: document.getElementById('current-folder-id'),
        parentFolderId: document.getElementById('parent-folder-id')
    };

    // Check if user is authenticated
    function checkAuth() {
        const token = localStorage.getItem('authToken');
        if (token) {
            showDashboard();
            // Add token to forms that need it
            document.querySelectorAll('form').forEach(form => {
                // Only add to forms that don't already have an auth input
                if (!form.querySelector('input[name="auth"]')) {
                    const authInput = document.createElement('input');
                    authInput.type = 'hidden';
                    authInput.name = 'auth';
                    authInput.value = token;
                    form.appendChild(authInput);
                }
            });
        } else {
            showAuth();
        }
    }

    // Show auth section
    function showAuth() {
        elements.authSection.classList.add('active-section');
        elements.dashboardSection.style.display = 'none';
    }

    // Show dashboard
    function showDashboard() {
        elements.authSection.classList.remove('active-section');
        elements.dashboardSection.style.display = 'block';
        const username = localStorage.getItem('username');
        if (username) {
            document.getElementById('username-display').textContent = username;
        }
    }

    // Switch between auth tabs
    function switchTab(tabType) {
        elements.tabs.forEach(tab => tab.classList.remove('active'));
        document.querySelector(`.tab[data-tab="${tabType}"]`).classList.add('active');
        
        document.querySelectorAll('.form-container').forEach(form => form.classList.remove('active-form'));
        document.querySelector(`.form-container.${tabType}`).classList.add('active-form');
    }

    // Toggle modal
    function toggleModal(modal, show = true) {
        if (show) {
            modal.classList.add('active');
        } else {
            modal.classList.remove('active');
        }
    }

    // Set folder ID for forms
    function setFolderId(id) {
        elements.currentFolderId.value = id;
        elements.parentFolderId.value = id;
    }

    // Event listeners
    function attachEventListeners() {
        // Auth tabs
        elements.tabs.forEach(tab => {
            tab.addEventListener('click', () => {
                const tabType = tab.getAttribute('data-tab');
                switchTab(tabType);
            });
        });

        // Intercept login form
        elements.loginForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            const username = document.getElementById('login-username').value;
            const password = document.getElementById('login-password').value;
            
            try {
                const response = await fetch('/auth/login', {
                    method: 'POST',
                    body: `${username}\n${password}`
                });
                
                if (!response.ok) {
                    throw new Error(await response.text());
                }
                
                const token = await response.text();
                localStorage.setItem('authToken', token);
                localStorage.setItem('username', username);
                
                // Redirect to dashboard
                window.location.reload();
                
            } catch (error) {
                document.getElementById('login-message').textContent = error.message;
                document.getElementById('login-message').className = 'message error';
            }
        });

        // Intercept register form
        elements.registerForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            const username = document.getElementById('register-username').value;
            const password = document.getElementById('register-password').value;
            const confirm = document.getElementById('register-confirm').value;
            
            if (password !== confirm) {
                document.getElementById('register-message').textContent = 'Passwords do not match';
                document.getElementById('register-message').className = 'message error';
                return;
            }
            
            try {
                const response = await fetch('/auth/register', {
                    method: 'POST',
                    body: `${username}\n${password}`
                });
                
                if (!response.ok) {
                    throw new Error(await response.text());
                }
                
                document.getElementById('register-message').textContent = 'Registration successful! You can now log in.';
                document.getElementById('register-message').className = 'message success';
                
                setTimeout(() => {
                    switchTab('login');
                }, 2000);
                
            } catch (error) {
                document.getElementById('register-message').textContent = error.message;
                document.getElementById('register-message').className = 'message error';
            }
        });

        // Toggle modals
        elements.uploadBtn.addEventListener('click', () => {
            toggleModal(elements.uploadModal);
        });
        
        elements.newFolderBtn.addEventListener('click', () => {
            toggleModal(elements.newFolderModal);
        });
        
        elements.closeButtons.forEach(btn => {
            btn.addEventListener('click', () => {
                const modal = btn.closest('.modal');
                toggleModal(modal, false);
            });
        });

        // Folder click listener - added dynamically when folders are loaded
        document.addEventListener('click', (e) => {
            if (e.target.classList.contains('folder') || e.target.parentElement.classList.contains('folder')) {
                const folderElement = e.target.classList.contains('folder') ? e.target : e.target.parentElement;
                const folderId = folderElement.getAttribute('data-id');
                if (folderId) {
                    setFolderId(folderId);
                    window.location.href = `/folder?id=${folderId}&auth=${localStorage.getItem('authToken')}`;
                }
            }
        });
    }

    // Initialize
    attachEventListeners();
    checkAuth();
});