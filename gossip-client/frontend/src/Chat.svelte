<script>
  import { onMount, createEventDispatcher } from 'svelte';
  import { ProgressRadial } from '@skeletonlabs/skeleton';
  import * as wails from '../wailsjs/runtime';
  import closeIcon from './assets/images/logout.svg';
  import xmark from './assets/images/close.svg';
  import hamburger from './assets/images/hamburger.svg';
  import settingsIcon from './assets/images/settings.svg';
  import user from './assets/images/user.svg';
  import Call from './components/Call.svelte';
  import Settings from './components/Settings.svelte';
  import { SendMessage, Disconnect } from '../wailsjs/go/main/App.js';
  import { marked } from 'marked';
  import { writable } from 'svelte/store';

  const dispatch = createEventDispatcher();
  let messageText = ''; // For new message input
  let expirySetting = "60"; // Default to 60 seconds
  let clientID = "n/a";
  let channels = [];
  let selectedChannel = ''; // To hold the currently selected channel
  let isOpen = writable(false);
  let callerList = {};

  // Function to toggle the sidebar
  function toggleSidebar() {
    isOpen.update(value => !value);
  }

  // Create a custom renderer
  const renderer = new marked.Renderer();
  const originalLinkRenderer = renderer.link;
  renderer.link = (href, title, text) => {
    const html = originalLinkRenderer.call(renderer, href, title, text);
    return html.replace(/^<a /, '<a target="_blank" ');
  };

  // Set the renderer to marked
  marked.setOptions({
    renderer
  });

  class ChatMessage {
    constructor(username, message, expiration, timestamp, channel, sender) {
      this.channel = channel;
      this.username = username;
      this.message = message;
      this.expiration = expiration;
      this.timestamp = timestamp;
      this.sender = sender;
    }
  }

  let isLoading = true;
  let loadingStatus = "Setting things up...";
  let showModal = false;
  let settingsFlag = false;
  let settings = null;
  let serverName = "Generic Gossip Server";

  export let username = 'You';

  let currentTime = Date.now();

  let messages = channels.map(channel => new ChatMessage("Server", `Welcome to the ${channel} channel!`, Math.floor(Date.now() / 1000) + 60, Math.floor(Date.now() / 1000), channel));

  /**
   * Formats the time since a message was sent
   * @param {number} timestamp - The timestamp of the message
   * @returns {string} - The formatted time since the message was sent
   */
  function formatTimeAgo(timestamp) {
    const secondsAgo = Math.floor((currentTime / 1000) - timestamp);
    if (secondsAgo < 10) {
      return `just now`;  
    } else if (secondsAgo < 60) {
      return `${secondsAgo} seconds ago`;
    } else if (secondsAgo < 3600) {
      return `${Math.floor(secondsAgo / 60)} minutes ago`;
    } else if (secondsAgo < 86400) {
      return `${Math.floor(secondsAgo / 3600)} hours ago`;
    } else {
      return `${Math.floor(secondsAgo / 86400)} days ago`;
    }
  }

  /**
   * Scrolls the chat messages container to the bottom.
   * This function should be called whenever a new message is added to the chat
   * to ensure that the latest message is visible to the user.
   * Now it specifically targets an element with the id 'bottom'.
   */
  function scrollToBottom() {
    const bottomElement = document.querySelector('#chat-messages');
    if (bottomElement) {
      setTimeout(() => {
        bottomElement.scrollTop = bottomElement.scrollHeight;
      }, 100);
    }
  }

  function changeChannel() {
    messages = [...messages];
  }

  function toggleModal() {
    if (settingsFlag == false) { settingsFlag = true }
    showModal = !showModal;
  }

  /**
   * Closes the chat and disconnects from the server
   */
  function closeChat() {
    Disconnect();
    messages = channels.map(channel => new ChatMessage("Server", `Welcome to the ${channel} channel!`, Math.floor(Date.now() / 1000) + 60, Math.floor(Date.now() / 1000), channel));
    
    dispatch('close');
  }

  /**
   * Gets the initials of a username
   * @param {string} name - The username
   * @returns {string} - The initials of the username
   */
  function getInitials(name) {
    return name.split(' ').map((n) => n[0]).join('');
  }

  /**
   * Dispatches a message to the server
   */
  function dispatchMessage() {
    if (messageText.trim() !== '') {
      SendMessage(messageText, Math.floor(Date.now() / 1000) + parseInt(expirySetting.toString()), selectedChannel);
      messageText = ''; // Clear input after sending
      scrollToBottom();
    }
  }

  onMount(() => {

    wails.EventsOn("update-client-id", (id) => {
      clientID = id;
    });

    wails.EventsOn("update-loading-status", (status) => {
      loadingStatus = status;
    });

    wails.EventsOn("finish-loading-status", () => {
      isLoading = false;
    });

    wails.EventsOn("server-name-received", (name) => {
      serverName = name;
    });

    wails.EventsOn("message-received", (channel, cUsername, cMessage, cExpiration, cTimestamp, cSender) => {
      let receivedMessage = new ChatMessage(cUsername, cMessage, cExpiration, cTimestamp, channel, cSender);
      receivedMessage.message = marked(receivedMessage.message); // Parse Markdown to HTML
      messages = [...messages, receivedMessage];
      scrollToBottom();
      
      if (callerList.hasOwnProperty(cSender)) {
        callerList[cSender] = cUsername;
      }
    });

    wails.EventsOn("channel-update", (channel) => {
      channels = [...channels, channel];
      // Automatically select the first channel from the list if available
      if (channels.length > 0) {
        selectedChannel = channels[0];
        changeChannel(); // Call changeChannel to update the chat based on the newly selected channel
      }
    });

    wails.EventsOn('caller_self_active', () => {
      if (!callerList.hasOwnProperty(clientID)) {
        callerList[clientID] = "You";
      }
    });

    wails.EventsOn("caller_self_hung_up", () => {
      if (callerList.hasOwnProperty(clientID)) {
        delete callerList[clientID];
      }
    });

    wails.EventsOn("caller_active", (callerID) => {
      if (!callerList.hasOwnProperty(callerID)) {
        callerList[callerID] = callerID; // Using callerID as username for now
      }
    });

    wails.EventsOn("caller_hung_up", (callerID) => {
      if (callerList.hasOwnProperty(callerID)) {
        delete callerList[callerID];
      }
    });

    setInterval(() => {
      currentTime = Date.now();
    }, 1000);
    setInterval(() => {
      expireMessages();
    }, 1000);
  });

  /**
   * Expires messages that have exceeded their expiration time
   */
  function expireMessages() {
    const currentTimeInSeconds = Math.floor(Date.now() / 1000);
    messages = messages.filter(message => currentTimeInSeconds <= message.expiration);
  }
</script>

<style>
  
  .white-icon {
    filter: brightness(0) invert(1);
  }

  .sidebar {
    width: 0;
    height: 100vh;
    position: fixed;
    right: 0;
    top: 0;
    overflow-x: hidden;
    transition: 0.5s;
    z-index: 1000;
  }

  .sidebar.open {
    width: 300px; /* Width of the sidebar when it's open */
  }
</style>

<div class="sidebar bg-surface-900 border-r border-surface-600 text-left" class:open={$isOpen}>
  <div class="relative w-full">
    <div class="w-[300px] min-w-[300px] px-5 py-3">
    
      <div class="w-full flex items-center justify-between">
        <div class="font-bold text-xl py-3">Connection Panel</div>
        <button class="absolute top-4 right-4 border-none" style="z-index:1001;" on:click={toggleSidebar}>
          <img alt="Close Icon" src="{xmark}" class="white-icon hover:scale-110 transition-all m-0 p-0" style="width: 24px; height: 24px;" draggable="false" />
        </button>
      </div>
    
      <hr class="opacity-60"/>
  
      <div class="font-bold py-3">Call List (Key Ring)</div>
  
      <div class="">
        {#each Object.entries(callerList) as [callerID, callerName]}
          <div class="flex items-center gap-4 pb-2">
            <div class="">
              <img alt="User Icon" src="{user}" class="white-icon hover:scale-110 transition-all m-0 p-0" style="width: 24px; height: 24px;" draggable="false" />
            </div>
            <div class="">{callerName.length > 10 ? `${callerName.substring(0, 10)}...` : callerName}</div>
          </div>
        {/each}
      </div>
    </div>
  </div>
</div>

<div class="absolute top-0 left-0 right-0 w-full bg-surface-700" style="z-index:50">
  <div class="w-full flex justify-between items-center">

    <div class="px-4 flex justify-center items-center">
      <span class="text-xl font-bold">{serverName}</span>
    </div>
    <select id="channel-select" bind:value={selectedChannel} on:change={changeChannel} class="channel-select bg-surface-900 border-2 border-surface-600 rounded-lg py-2 focus:ring-0 focus:border-surface-600 w-1/3 m-1 mt-2" style="text-align-last: center;">
      {#each channels as channel}
        <option value={channel}>{channel}</option>
      {/each}
    </select>
    <div class="flex justify-evenly items-center gap-4 pr-6">
      
        <button on:click={closeChat} class="cursor-pointer border-none">
          <img alt="Close Icon" src="{closeIcon}" class="white-icon hover:scale-110 transition-all m-0" style="width: 24px; height: 24px;" draggable="false" />
        </button>

        <button on:click={toggleModal} class="cursor-pointer border-none">
          <img alt="Settings Icon" src="{settingsIcon}" class="white-icon hover:scale-110 transition-all m-0" style="width: 24px; height: 24px;" draggable="false" />
        </button>

        <button on:click={toggleSidebar} class="cursor-pointer border-none">
            <img alt="Menu Icon" src="{hamburger}" class="white-icon hover:scale-110 transition-all m-0" style="width: 24px; height: 24px;" draggable="false" />
        </button>

    </div>
  </div>

  <div class="bg-surface-500">
    <Call />
  </div>

</div>
{#if isLoading}
  <div class="loader">{loadingStatus}</div>
  <ProgressRadial value={undefined} class="scale-50"/>
{:else}
  <div class="chat-container w-full absolute bottom-0">
    <div id="chat-messages" class="overflow-y-scroll max-h-[90vh] p-4 pt-24">
      {#each messages as message}
      {#if selectedChannel === message.channel}
      <div class="message {message.sender === clientID ? 'from-user' : ''} w-full">
        <div class="flex {message.sender === clientID ? 'flex-row-reverse' : ''} w-full mt-6">
          <div class="mx-3 inline-block initials rounded-full w-8 h-8 flex items-center justify-center border-2 border-surface-200 p-4 uppercase">
            {getInitials(message.username)}
          </div>

          <div class="{message.sender === clientID ? 'rounded-tr-none bg-primary-900' : 'rounded-tl-none bg-surface-700'}  rounded-lg px-5 py-3 w-full max-w-[50vw] md:max-w-[40vw]">
            <div class="flex justify-between">
              <span class="font-bold">{message.username}</span>
              <span class="flex items-center">
                <span class="opacity-60 pr-2">{formatTimeAgo(message.timestamp)}</span>
                <span title={`Expires in ${Math.round((message.expiration - currentTime / 1000) / 60)} minute(s)`}>
                  <ProgressRadial value={Math.max(0, Math.min(100, ((message.expiration - currentTime / 1000) / (message.expiration - message.timestamp)) * 100))} stroke={75} class="w-[18px] h-[18px]"/>
                </span>
              </span>
            </div>
            <div class="block text-left w-full whitespace-pre-wrap break-words -mb-4">
              {@html message.message}
            </div>
          </div>
          
        </div>
      </div>
      {/if}
      {/each}
      <div id="bottom"></div>
    </div>
    <div class="flex p-4 pt-2 gap-4">
      <textarea bind:value={messageText} placeholder="Type a message..." class="input px-4 py-3 border-surface-700 focus:outline-none focus:ring-0 rounded-lg resize-none" rows="1" maxlength="15000"
        on:keydown={(event) => {
          if (event.key === 'Enter') {
            if (event.ctrlKey) {
              // Add a new line if Ctrl+Enter is pressed
              messageText += '\n';
            } else if (event.shiftKey) {
              // Prevent default behavior to avoid adding a new line in the textarea
            } else {
              event.preventDefault();
              // Dispatch the message if only Enter is pressed
              dispatchMessage();
            }
          }
        }} />

      <select id="expiry-select" bind:value={expirySetting} class="channel-select bg-surface-900 border-2 border-surface-600 rounded-lg px-4 py-3 focus:ring-0 focus:border-surface-600 w-32 text-sm" placeholder="60s">
        <option value="30">30s</option>
        <option value="60" selected>60s</option>
        <option value="300">5m</option>
        <option value="900">15m</option>
        <option value="1800">30m</option>
        <option value="3600">1h</option>
        <option value="18000">5h</option>
        <option value="62400">24h</option>
      </select>
      <button on:click={dispatchMessage} type="submit" class="btn variant-filled-primary rounded-lg px-4 py-3">
        Send
      </button>
    </div>
  </div>

  <!-- Add the Settings component -->
  <Settings showModal={showModal} onClose={toggleModal} bind:settings={settings} />
{/if}