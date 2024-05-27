import Toast from '../components/Toast.svelte';

/**
 * Function to create a toast notification.
 * @param {string} message - The message to display in the toast.
 * @param {number} duration - Duration for which the toast should be visible.
 */
export function createToast(message, duration = 3000) {
    const toastContainer = document.createElement('button');
    document.body.appendChild(toastContainer);

    const toast = new Toast({
        target: toastContainer,
        props: {
            message,
            timeout: duration
        }
    });

    // Clean up the toast element after it's done
    toast.$on('destroy', () => {
        document.body.removeChild(toastContainer);
    });
}