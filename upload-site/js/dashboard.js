document.addEventListener('DOMContentLoaded', () => {
    // DOM elements
    const elements = {
        dashboardSection: document.getElementById('dashboard-section'),
        uploadBtn: document.getElementById('upload-btn'),
        uploadModal: document.getElementById('upload-modal'),
        newFolderBtn: document.getElementById('new-folder-btn'),
        newFolderModal: document.getElementById('new-folder-modal'),
        newInventoryBtn: document.getElementById('new-inventory-btn'),
        newInventoryModal: document.getElementById('new-inventory-modal'),
        deleteConfirmModal: document.getElementById('delete-confirm-modal'),
        inventoryTree: document.getElementById('inventory-tree'),
        folderTree: document.getElementById('folder-tree'),
        itemsContainer: document.getElementById('items-container'),
        breadcrumbs: document.getElementById('breadcrumbs'),
        closeButtons: document.querySelectorAll('.close'),
        currentFolderId: document.getElementById('current-folder-id'),
        parentFolderId: document.getElementById('parent-folder-id'),
        fileUpload: document.getElementById('file-upload'),
        fileName: document.getElementById('file-name'),
        uploadPreview: document.getElementById('upload-preview'),
        previewFilename: document.getElementById('preview-filename'),
        previewFilesize: document.getElementById('preview-filesize'),
        usernameDisplay: document.getElementById('username-display')
    };

    // Global variables
    let currentInventoryId = null;
    let currentFolderId = null;

    // Check if user is authenticated
    function checkAuth() {
        const token = localStorage.getItem('authToken');
        if (!token) {
            // User is not logged in, redirect to login page
            window.location.href = '/login';
            return;
        }
        
        // Set username in header
        const username = localStorage.getItem('username');
        if (username && elements.usernameDisplay) {
            elements.usernameDisplay.textContent = username;
        }
        
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
        
        // Load inventories
        loadInventories();
    }

    // Format file size for display
    function formatFileSize(bytes) {
        if (bytes < 1024) return bytes + ' bytes';
        else if (bytes < 1048576) return (bytes / 1024).toFixed(1) + ' KB';
        else return (bytes / 1048576).toFixed(1) + ' MB';
    }
    
    // Error handling for API requests
    function handleApiError(error, containerElement, message = 'Failed to load data') {
        console.error(error);
        if (containerElement) {
            containerElement.innerHTML = `<p class="empty-folder">${message}</p>`;
        }
    }

    // Load user's inventories from server
    async function loadInventories() {
        if (!elements.inventoryTree) return;
        
        try {
            elements.inventoryTree.innerHTML = '<div class="loading-spinner"></div>';
            
            const token = localStorage.getItem('authToken');
            const response = await fetch(`/api/inventories?auth=${token}`);
            
            if (!response.ok) {
                throw new Error(`Failed to load inventories: ${response.status}`);
            }
            
            const data = await response.json();
            
            if (!data.success) {
                throw new Error('Failed to load inventories');
            }
            
            elements.inventoryTree.innerHTML = '';
            
            if (data.data.length === 0) {
                elements.inventoryTree.innerHTML = '<p class="empty-folder">No inventories found</p>';
                return;
            }
            
            data.data.forEach(inventory => {
                const inventoryElement = document.createElement('div');
                inventoryElement.className = 'inventory';
                inventoryElement.dataset.id = inventory.id;
                inventoryElement.dataset.rootFolderId = inventory.rootFolderId;
                inventoryElement.innerHTML = `<i class="fas fa-box"></i> ${inventory.name}`;
                
                inventoryElement.addEventListener('click', () => {
                    currentInventoryId = inventory.id;
                    loadRootFolder(inventory.id);
                    
                    // Highlight the selected inventory
                    document.querySelectorAll('.inventory').forEach(el => {
                        el.classList.remove('active');
                    });
                    inventoryElement.classList.add('active');
                });
                
                elements.inventoryTree.appendChild(inventoryElement);
            });
            
        } catch (error) {
            handleApiError(error, elements.inventoryTree, 'Failed to load inventories');
        }
    }

    // Load root folder of an inventory
    async function loadRootFolder(inventoryId) {
        try {
            const token = localStorage.getItem('authToken');
            
            // Use the inventory data that should already include rootFolderId
            const inventories = document.querySelectorAll('.inventory');
            let rootFolderId = null;
            
            // First check if we already have the rootFolderId from the inventory list
            inventories.forEach(inv => {
                if (parseInt(inv.dataset.id) === inventoryId && inv.dataset.rootFolderId) {
                    rootFolderId = parseInt(inv.dataset.rootFolderId);
                }
            });
            
            // If not found in DOM, fetch it
            if (!rootFolderId) {
                const response = await fetch(`/api/inventory/rootFolder?inventoryId=${inventoryId}&auth=${token}`);
                
                if (!response.ok) {
                    throw new Error(`Failed to get root folder: ${response.status}`);
                }
                
                const data = await response.json();
                if (!data.success) {
                    throw new Error('Failed to get root folder');
                }
                
                rootFolderId = data.rootFolderId;
            }
            
            // Load the folder contents with the actual root folder ID
            loadFolderContents(rootFolderId);
        } catch (error) {
            console.error('Error loading root folder:', error);
        }
    }

    // Toggle modal
    function toggleModal(modal, show = true) {
        if (modal) {
            if (show) {
                modal.classList.add('active');
            } else {
                modal.classList.remove('active');
            }
        }
    }

    // Event listeners
    function attachEventListeners() {
        // Open modals
        if (elements.uploadBtn && elements.uploadModal) {
            elements.uploadBtn.addEventListener('click', () => {
                toggleModal(elements.uploadModal, true);
            });
        }
        
        if (elements.newFolderBtn && elements.newFolderModal) {
            elements.newFolderBtn.addEventListener('click', () => {
                toggleModal(elements.newFolderModal, true);
            });
        }
        
        if (elements.newInventoryBtn && elements.newInventoryModal) {
            elements.newInventoryBtn.addEventListener('click', () => {
                toggleModal(elements.newInventoryModal, true);
            });
        }
        
        // Close modals
        if (elements.closeButtons) {
            elements.closeButtons.forEach(btn => {
                btn.addEventListener('click', () => {
                    const modal = btn.closest('.modal');
                    toggleModal(modal, false);
                });
            });
        }

        // File upload preview
        if (elements.fileUpload) {
            elements.fileUpload.addEventListener('change', (e) => {
                const file = e.target.files[0];
                if (file) {
                    if (elements.fileName) {
                        elements.fileName.textContent = file.name;
                    }
                    if (elements.previewFilename) {
                        elements.previewFilename.textContent = file.name;
                    }
                    if (elements.previewFilesize) {
                        elements.previewFilesize.textContent = formatFileSize(file.size);
                    }
                    if (elements.uploadPreview) {
                        elements.uploadPreview.classList.remove('hidden');
                    }
                } else {
                    if (elements.fileName) {
                        elements.fileName.textContent = 'Choose a file...';
                    }
                    if (elements.uploadPreview) {
                        elements.uploadPreview.classList.add('hidden');
                    }
                }
            });
        }
        
        // Form submission handlers
        // New folder form
        const newFolderForm = document.getElementById('new-folder-form');
        if (newFolderForm) {
            newFolderForm.addEventListener('submit', async (e) => {
                e.preventDefault();
                
                const folderName = newFolderForm.querySelector('#folder-name').value;
                const folderId = newFolderForm.querySelector('input[name="folderId"]').value || currentFolderId || 1;
                const authToken = localStorage.getItem('authToken');
                
                if (!folderName) {
                    alert('Please enter a folder name');
                    return;
                }
                
                try {
                    const response = await fetch(`/addFolder?folderName=${encodeURIComponent(folderName)}&folderId=${folderId}&auth=${authToken}`, {
                        method: 'GET'
                    });
                    
                    if (!response.ok) {
                        throw new Error(await response.text() || 'Failed to create folder');
                    }
                    
                    // Close modal and redirect to the new folder
                    toggleModal(elements.newFolderModal, false);
                    
                    // Get the new folder ID from the response if possible
                    // Otherwise, reload the current folder
                    try {
                        const result = await response.json();
                        if (result && result.folderId) {
                            window.location.href = `/folder?id=${result.folderId}&auth=${authToken}`;
                        } else {
                            loadFolderContents(folderId);
                        }
                    } catch (e) {
                        // If can't parse JSON, just reload current folder
                        window.location.href = `/folder?id=${folderId}&auth=${authToken}`;
                    }
                    
                    // Clear form
                    newFolderForm.reset();
                    
                } catch (error) {
                    console.error('Error creating folder:', error);
                    alert('Failed to create folder: ' + error.message);
                }
            });
        }
        
        // New inventory form
        const newInventoryForm = document.getElementById('new-inventory-form');
        if (newInventoryForm) {
            newInventoryForm.addEventListener('submit', async (e) => {
                e.preventDefault();
                
                const inventoryName = newInventoryForm.querySelector('#inventory-name').value;
                const authToken = localStorage.getItem('authToken');
                
                if (!inventoryName) {
                    alert('Please enter an inventory name');
                    return;
                }
                
                try {
                    const response = await fetch(`/addInventory?inventoryName=${encodeURIComponent(inventoryName)}&auth=${authToken}`, {
                        method: 'POST'
                    });
                    
                    if (!response.ok) {
                        throw new Error(await response.text() || 'Failed to create inventory');
                    }
                    
                    // Close modal and reload inventories
                    toggleModal(elements.newInventoryModal, false);
                    loadInventories();
                    
                    // Clear form
                    newInventoryForm.reset();
                    
                } catch (error) {
                    console.error('Error creating inventory:', error);
                    alert('Failed to create inventory: ' + error.message);
                }
            });
        }
        
        // Upload form
        const uploadForm = document.getElementById('upload-form');
        if (uploadForm) {
            uploadForm.addEventListener('submit', async (e) => {
                e.preventDefault();
                
                const fileInput = uploadForm.querySelector('#file-upload');
                const folderId = uploadForm.querySelector('input[name="folderId"]').value || currentFolderId || 1;
                const authToken = localStorage.getItem('authToken');
                
                if (!fileInput.files || fileInput.files.length === 0) {
                    alert('Please select a file to upload');
                    return;
                }
                
                // Show progress bar
                const progressContainer = document.getElementById('progress-container');
                const progressBar = document.getElementById('progress-bar');
                
                if (progressContainer) progressContainer.classList.remove('hidden');
                if (progressBar) progressBar.style.width = '0%';
                
                try {
                    const formData = new FormData(uploadForm);
                    
                    // Create XHR for progress tracking
                    const xhr = new XMLHttpRequest();
                    xhr.open('POST', `/upload?folderId=${folderId}&auth=${authToken}`);
                    
                    xhr.upload.addEventListener('progress', (event) => {
                        if (event.lengthComputable && progressBar) {
                            const percentComplete = (event.loaded / event.total) * 100;
                            progressBar.style.width = percentComplete + '%';
                        }
                    });
                    
                    // Set up promise to handle response
                    const uploadPromise = new Promise((resolve, reject) => {
                        xhr.onload = () => {
                            if (xhr.status >= 200 && xhr.status < 300) {
                                resolve(xhr.responseText);
                            } else {
                                reject(new Error(xhr.statusText || xhr.responseText || 'Upload failed'));
                            }
                        };
                        xhr.onerror = () => reject(new Error('Network error during upload'));
                    });
                    
                    // Send the data
                    xhr.send(formData);
                    
                    // Wait for completion
                    await uploadPromise;
                    
                    // Complete the progress bar
                    if (progressBar) progressBar.style.width = '100%';
                    
                    // Close modal and redirect to folder view after a short delay
                    setTimeout(() => {
                        toggleModal(elements.uploadModal, false);
                        window.location.href = `/folder?id=${folderId}&auth=${authToken}`;
                        
                        // Reset form and progress
                        uploadForm.reset();
                        if (elements.fileName) elements.fileName.textContent = 'Choose a file...';
                        if (elements.uploadPreview) elements.uploadPreview.classList.add('hidden');
                        if (progressContainer) progressContainer.classList.add('hidden');
                    }, 1000);
                    
                } catch (error) {
                    console.error('Error uploading file:', error);
                    alert('Failed to upload file: ' + error.message);
                    
                    // Hide progress
                    if (progressContainer) progressContainer.classList.add('hidden');
                }
            });
        }
    }

    // Initialize the page
    attachEventListeners();
    checkAuth();
});
