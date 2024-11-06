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

