<script>
    import { onMount } from 'svelte';
    import closeIcon from '../assets/images/close.svg';
    import { LoadSettings, SaveSettings } from '../../wailsjs/go/main/App.js';
  
    export let showModal = false; // Prop to control the visibility of the modal
    export let onClose; // Prop to handle the close event
  
    export let settings; // Initialize settings as undefined
  
    onMount(async () => {
      const loadedSettings = await LoadSettings();
      // Set settings with loaded values or default values if none are loaded
      settings = loadedSettings || {
        selectedTheme: 'wintry',
        defaultUsername: '',
        defaultHost: '',
        defaultPort: '1720',
      };
      setTimeout(updateTheme, 100);
    });
  
    /**
     * Updates the settings and ensures the settings variable is updated
     * @returns {Promise<void>}
     */
    async function updateSettings() {
        await SaveSettings(settings);
    }
  
    /**
     * Updates the theme on the document.body element based on the selected theme
     */
    function updateTheme() {
      if (settings && settings.selectedTheme) {
        document.body.setAttribute('data-theme', settings.selectedTheme);
      }
    }
  
    /**
     * Reactive statement to call updateSettings when settings change
     */
    $: if (settings) {
      setTimeout(updateSettings, 100);
    }
    $: if (settings && settings.selectedTheme) {
      setTimeout(updateTheme, 100);
    }
  
  </script>
  
  {#if showModal && settings}
    <div class="absolute inset-0 flex items-center justify-center bg-black bg-opacity-50" style="z-index:60;">
      <div class="w-full h-full bg-surface-900 p-4 text-white flex flex-col">
        <button class="absolute top-4 right-4 cursor-pointer" on:click={onClose}>
          <img alt="Close Icon" src="{closeIcon}" class="white-icon hover:scale-110 transition-all" style="width: 24px; height: 24px;">
        </button>
  
        <div class="text-left ">
          <h2 class="text-2xl font-bold">Settings</h2>
          <hr class="opacity-50 mt-6" />
        </div>
  
        <div class="max-h-[80vh] overflow-y-scroll w-full pt-4">
          <div class="flex items-center justify-between w-full p-4 mx-auto max-w-[400px]">
            <label for="theme-select" class="block text-lg font-medium mr-4">Theme</label>
            <select id="theme-select" bind:value={settings.selectedTheme} class="bg-surface-900 border-2 border-surface-700 rounded-lg p-2 w-1/2 focus:ring-0 focus:border-surface-700" style="text-align-last: center;">
              <option value="wintry">Wintry</option>
              <option value="crimson">Crimson</option>
              <option value="vintage">Vintage</option>
              <option value="seafoam">Seafoam</option>
              <option value="modern">Modern</option>
              <option value="rocket">Rocket</option>
              <option value="skeleton">Skeleton</option>
              <option value="sahara">Sahara</option>
              <option value="hamlindigo">Hamlindigo</option>
              <option value="gold-nouveau">Gold Nouveau</option>
            </select>
          </div>
  
          <hr class="opacity-70 py-2 w-full p-4 mx-auto max-w-[400px] mt-4" />
  
          <div class="flex items-center justify-between w-full p-4 mx-auto max-w-[400px]">
            <label for="default-username" class="block text-lg font-medium mr-4">Default Username</label>
            <input id="default-username" bind:value={settings.defaultUsername} placeholder="Set default username" class="bg-surface-900 border-2 border-surface-700 rounded-lg p-2 w-1/2 focus:ring-0 focus:border-surface-700" />
          </div>
  
          <hr class="opacity-70 py-2 w-full p-4 mx-auto max-w-[400px] mt-4" />
  
          <div class="flex items-center justify-between w-full p-4 mx-auto max-w-[400px]">
            <label for="default-host" class="block text-lg font-medium mr-4">Default Host</label>
            <input id="default-host" bind:value={settings.defaultHost} placeholder="Set default host" class="bg-surface-900 border-2 border-surface-700 rounded-lg p-2 w-1/2 focus:ring-0 focus:border-surface-700" />
          </div>
  
          <hr class="opacity-70 py-2 w-full p-4 mx-auto max-w-[400px] mt-4" />
  
          <div class="flex items-center justify-between w-full p-4 mx-auto max-w-[400px]">
            <label for="default-port" class="block text-lg font-medium mr-4">Default Port</label>
            <input id="default-port" bind:value={settings.defaultPort} placeholder="Set default port" class="bg-surface-900 border-2 border-surface-700 rounded-lg p-2 w-1/2 focus:ring-0 focus:border-surface-700" />
          </div>
        </div>
      </div>
    </div>
{/if}

<style>
  .white-icon {
    filter: brightness(0) invert(1);
  }
</style>