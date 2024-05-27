<script>
    import { onMount } from 'svelte';
    import { fade } from 'svelte/transition';

    export let message = '';
    export let timeout = 3000; // default timeout in milliseconds

    let show = false;

    // Function to show the toast
    function showToast() {
        show = true;
        setTimeout(() => {
            show = false;
        }, timeout);
    }

    onMount(() => {
        showToast();
    });
</script>

{#if show}
    <button transition:fade class="fixed top-32 right-4 bg-red-700 text-white p-4 rounded-lg shadow-lg flex justify-center items-center gap-2" on:click={() => show = false}>
        <span class="text-4xl font-bold">&#9888;</span>{message}
    </button>
{/if}