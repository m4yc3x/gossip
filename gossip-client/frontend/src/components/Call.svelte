<script>
  import { UpdateCallID, ToggleGoMute, ToggleGoDeaf, StartRecording, StopRecording } from '../../wailsjs/go/main/App.js';
  import callIcon from '../assets/images/call.svg';
  import hangupIcon from '../assets/images/hangup.svg';
  import { onMount } from 'svelte';
  import * as wails from '../../wailsjs/runtime';
  import { createToast } from './toast';
  import micIcon from '../assets/images/microphone.svg'
  import headphoneIcon from '../assets/images/headphones.svg'

  let inCall = false;
  let muted = false;
  let deafened = false;
  let callStatus = "Not in Call";
  let callStartTime;
  let interval;
  let callID = "";

  function start() {
    inCall = true;
    StartRecording();
  }
  
  function stop() {
    inCall = false;
    StopRecording();
  }

  /**
   * Toggles the mute state of the microphone.
   * If the microphone is currently muted, it will unmute it, and vice versa.
   */
  function toggleMute() {
    muted = !muted;
    ToggleGoMute();
    // Emit an event to the backend to handle the actual muting logic in the audio stream
  }

  /**
   * Toggles the mute state of the microphone.
   * If the microphone is currently muted, it will unmute it, and vice versa.
   */
   function toggleDeafen() {
    deafened = !deafened;
    ToggleGoDeaf();
    // Emit an event to the backend to handle the actual muting logic in the audio stream
  }

  function updateCallTime() {
    const now = new Date();
    const elapsed = new Date(now - callStartTime);
    const hours = elapsed.getUTCHours().toString().padStart(2, '0');
    const minutes = elapsed.getUTCMinutes().toString().padStart(2, '0');
    const seconds = elapsed.getUTCSeconds().toString().padStart(2, '0');
    callStatus = `Call started - ${hours}:${minutes}:${seconds}`;
  }

  $: UpdateCallID(callID);

  onMount(() => {
    wails.EventsOn("no-call-keys", () => {
    inCall = false;
      createToast('Your call list is empty', 7000);
    });

    wails.EventsOn("audio-activity", (dest, activity, exp, tms, snd) => {
      // TODO voice activity
    });

    wails.EventsOn("transfer-client", (uid, name) => {
      // TODO voice activity
    });

    wails.EventsOn("call_starting", () => {
      callStatus = "Setting up call...";
    });

    wails.EventsOn("call_active", (callID) => {
      callStatus = "Awaiting peers... Call ID: " + callID;
    });

    wails.EventsOn("call_not_found", (callID) => {
      stop();
      callStatus = "Call not found: " + callID;
    });

    wails.EventsOn("call_sending_offer", () => {
      callStatus = "Sending offer...";
    });

    wails.EventsOn("call_received_offer", () => {
      callStatus = "Received offer, sending answer...";
    });

    wails.EventsOn("call_received_answer", () => {
      callStatus = "Answer received!";
    });

    wails.EventsOn("call_received_ice", () => {
      callStatus = "Received ICE candidate.";
    });

    wails.EventsOn("call_started", () => {
      callStartTime = new Date();
      interval = setInterval(updateCallTime, 1000);
      callStatus = "Call started - 00:00:00";
    });

    wails.EventsOn("hang-up", () => {
      callStatus = "Call ended";
      clearInterval(interval);
      inCall = false;
    });
  });
</script>

<div class="w-full flex justify-between items-center">
  <div class="px-5 select-text">{callStatus}</div>
  <div class="">
    <input type="text" bind:value={callID} placeholder="Enter a call ID" class="outline-none" />
  </div>
  <div class="flex justify-center items-center gap-4 px-5 py-3 flex-row-reverse">
    {#if inCall}
      <button on:click={stop}>
        <img alt="Hangup Icon" src="{hangupIcon}" class="red-icon hover:scale-110 transition-all m-0 rotate-90" style="width: 24px; height: 24px; stroke: #ff0000;" draggable="false" title="Hangup"/>
      </button>
    {:else}
      <button on:click={start}>
        <img alt="Call Icon" src="{callIcon}" class="white-icon hover:scale-110 transition-all m-0" style="width: 24px; height: 24px; stroke: #fff;" draggable="false" title="Call"/>
      </button>
    {/if}
    {#if muted}
      <button on:click={toggleMute}>
        <img alt="Mic Icon" src="{micIcon}" class="red-icon hover:scale-110 transition-all m-0" style="width: 24px; height: 24px; stroke: #fff; stroke-width: 2;" draggable="false" title="Unmute"/>
      </button>
    {:else}
      <button on:click={toggleMute}>
        <img alt="Mic Icon" src="{micIcon}" class="white-icon hover:scale-110 transition-all m-0" style="width: 24px; height: 24px; stroke: #fff; stroke-width: 2;" draggable="false" title="Mute"/>
      </button>
    {/if}
    {#if deafened}
    <button on:click={toggleDeafen}>
      <img alt="Headphone Icon" src="{headphoneIcon}" class="red-icon hover:scale-110 transition-all m-0" style="width: 24px; height: 24px; stroke: #fff;" draggable="false" title="Undeafen"/>
    </button>
  {:else}
    <button on:click={toggleDeafen}>
      <img alt="Headphone Icon" src="{headphoneIcon}" class="white-icon hover:scale-110 transition-all m-0" style="width: 24px; height: 24px; stroke: #fff;" draggable="false" title="Deafen"/>
    </button>
  {/if}
  </div>
</div>

<style>
  .white-icon {
    filter: brightness(0) invert(1);
  }
  .red-icon {
    filter: brightness(0) saturate(100%) invert(13%) sepia(99%) saturate(7477%) hue-rotate(356deg) brightness(100%) contrast(119%);
  }
</style>