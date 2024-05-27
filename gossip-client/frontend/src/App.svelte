<script>
  import logo from './assets/images/logo-universal.png';
  import settingsIcon from './assets/images/settings.svg';
  import * as wails from '../wailsjs/runtime';
  import { LoadSettings, Boot } from '../wailsjs/go/main/App.js';
  import { onMount } from 'svelte';
  import Chat from './Chat.svelte';
  import Settings from './components/Settings.svelte';

  let host = '';
  let port = '1720';
  let password = '';
  let username = '';
  let showModal = false;
  let showChat = false;
  let connectionError = false;
  let passwordError = false;
  let settings = null;

  /**
   * Lifecycle function that runs when the component is mounted
   */
  onMount(async () => {
    const settings = await LoadSettings();
    host = settings.defaultHost;
    username = settings.defaultUsername;
    port = settings.defaultPort;
    document.body.setAttribute('data-theme', settings.selectedTheme);

    wails.EventsOn("server-disconnect", () => {
      showChat = false;
      connectionError = true;
    });
    wails.EventsOn("unauthorized", () => {
      showChat = false;
      passwordError = true;
    });
  });

  /**
   * Starts the chat session
   */
  function start() {
    Boot(host, parseInt(port), username, password);
    showChat = true;
  }

  /**
   * Handles the form submission
   * @param {Event} event - The form submission event
   */
  function handleSubmit(event) {
    event.preventDefault();
    start();
  }

  /**
   * Toggles the settings modal
   */
  function toggleModal() {
    showModal = !showModal;
  }

  /**
   * Handles the chat close event
   */
  function handleChatClose() {
    showChat = false;
  }
</script>

<main class="select-none flex flex-col items-center justify-center h-screen relative">
  {#if !showChat}
    <button on:click={toggleModal} class="absolute top-4 right-4 cursor-pointer border-none">
      <img alt="Settings Icon" src="{settingsIcon}" class="white-icon hover:scale-110 transition-all" style="width: 24px; height: 24px;" draggable="false" />
    </button>

    <img alt="Gossip logo" id="logo" src="{logo}" class="w-1/2 h-40 object-contain md:h-60 xl:h-80 mx-auto" draggable="false">
    
    <h1 class="text-5xl md:text-6xl text-center font-bold">GOSSIP</h1>
    <form on:submit={handleSubmit} class="container mx-auto p-8 space-y-8 mx-auto max-w-[400px]">
      <div class="space-y-4">

        {#if connectionError}
        <span class="text-error-500">Could not connect to server</span>
        {/if}
        {#if passwordError}
        <span class="text-error-500">Could not authenticate with server</span>
        {/if}
        <div class="grid grid-cols-[70%,1fr] gap-4">
          <input id="host" bind:value={host} placeholder="Enter server host" class="input px-4 py-3 focus:outline-none focus:ring-0 rounded-lg" required/>
          <input id="port" bind:value={port} placeholder="Port" class="input px-4 py-3 focus:outline-none focus:ring-0 rounded-lg" required maxlength="5"/>
        </div>
        <div>
          <input id="password" type="password" bind:value={password} placeholder="Enter password" class="input px-4 py-3 focus:outline-none focus:ring-0 rounded-lg" maxlength="500" />
        </div>
        <div>
          <input id="username" bind:value={username} placeholder="Enter username" class="input px-4 py-3 focus:outline-none focus:ring-0 rounded-lg" maxlength="25" pattern="^[a-zA-Z0-9_]+$" title="Username must be alphanumeric and can contain underscores" required/>
        </div>
        <button type="submit" class="btn variant-filled-primary w-full rounded-lg px-4 py-3">Connect</button>
      </div>
    </form>
  {:else}
    <Chat on:close="{handleChatClose}" username="{username}" /> <!-- Display the Chat component when showChat is true, passing the username -->
  {/if}

  <!-- Add the Settings component -->
  <Settings showModal={showModal} onClose={toggleModal} bind:settings={settings} />
</main>

<style>
  .white-icon {
    filter: brightness(0) invert(1);
  }
</style>