// Debounce function to limit the frequency of form submission
function debounce(func, delay) {
    let debounceTimer;
    return function() {
        const context = this;
        const args = arguments;
        clearTimeout(debounceTimer);
        debounceTimer = setTimeout(() => func.apply(context, args), delay);
    };
}

// Auto-save or delete if empty function with debounce
const debouncedSave = debounce((inputElement) => {
    // Check if the task title is empty
    if (inputElement.value.trim() === "") {
        // Change form action to delete if title is empty
        inputElement.form.action = `/delete?id=${inputElement.form.querySelector('input[name="id"]').value}`;
    }
    // Submit the form either for saving or deletion
    inputElement.form.submit();
}, 1000); // Save after 1000ms of no typing


async function handleLogin(event) {
    event.preventDefault();
    
    // Clear previous error messages
    document.getElementById('usernameError').textContent = '';
    document.getElementById('passwordError').textContent = '';
    
    const formData = new FormData(event.target);
    
    try {
        const response = await fetch('/login', {
            method: 'POST',
            body: formData
        });
        
        const data = await response.json();
        
        if (response.ok) {
            // Redirect on successful login
            window.location.href = data.redirect;
        } else {
            // Display error message in the appropriate error span
            if (data.field === 'username') {
                document.getElementById('usernameError').textContent = data.message;
            } else if (data.field === 'password') {
                document.getElementById('passwordError').textContent = data.message;
            } else {
                alert(data.message); // For other types of errors (if any)
            }
        }
    } catch (error) {
        console.error('Login error:', error);
        alert('An unexpected error occurred. Please try again.');
    }
}

async function handleRegister(event) {
    event.preventDefault(); // Prevent form from submitting normally
    const form = event.target;

    // Clear previous error messages
    document.getElementById('usernameError').textContent = "";
    document.getElementById('passwordError').textContent = "";
    document.getElementById('confirmPasswordError').textContent = "";

    // Collect form data
    const formData = new FormData(form);
    const password = formData.get('password');
    const confirmPassword = formData.get('confirmPassword');

    if (password !== confirmPassword) {
        document.getElementById('confirmPasswordError').textContent = "Passwords do not match.";
        return;
}
    const response = await fetch('/register', {
        method: 'POST',
        body: formData
    });

    const result = await response.json();

    // Check if registration was successful
    if (response.ok) {
        window.location.href = '/login';
    } else {
        // Display error message under the relevant field
        if (result.field === "username") {
            document.getElementById('usernameError').textContent = result.message;
        } else if (result.field === "password") {
            document.getElementById('passwordError').textContent = result.message;
        } else if (result.field === "confirmPassword") {
          document.getElementById('confirmPasswordError').textContent = result.message;
        } else {
            alert(result.message); // General error
        }
    }
}

// // Add event listener when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    document.getElementById('loginForm').addEventListener('submit', handleLogin);
});