<script lang="ts">
  import IconButton from "@smui/icon-button";
  import { onMount } from "svelte";
  import Progress from "../../components/Progress.svelte";
  import { createSnackbar } from "../../global";
  import { reminders } from "./main";

  export let id: number;
  export let name: string;
  export let description: string;
  export let priority: number;
  export let dueDate: string;
  export let createdDate: string;
  export let userWasNotified: boolean;

  let loading = false;
  let deleted = false;

  let priorityColor: string;
  const priorities = ["Low", "Normal", "Medium", "High", "Urgent"];

  onMount(() => {
    switch (priority) {
      case 0:
        priorityColor = "#707070";
        break;
      case 1:
        priorityColor = "#00ff00";
        break;
      case 2:
        priorityColor = "#0000ff";
        break;
      default:
        priorityColor = "#ff00ff";
        break;
    }
  });

  async function deleteSelf() {
    loading = true;
    try {
      const res = await (
        await fetch("/api/reminder/delete", {
          headers: { "Content-Type": "application/json" },
          method: "DELETE",
          body: JSON.stringify({ id }),
        })
      ).json();
      if (!res.success) throw Error();
      deleted = true;
      setTimeout(() => {
        $reminders = $reminders.filter(n => n.id !== id)
      }, 300)
    } catch (err) {
      $createSnackbar("Could not mark reminder as completed");
    }
    loading = false;
  }

  let container: HTMLDivElement;
  $: if (deleted) {
    container.style.setProperty(
      "--height",
      container.getBoundingClientRect().height + "px"
    );
    container.getBoundingClientRect();
    container.style.height = "0";
  }
</script>

<div
  bind:this={container}
  class="root mdc-elevation--z3"
  class:deleted
  style:--clr-priority={priorityColor}
>
  <div id="top">
    <h6>{name}</h6>
    <div id="buttons">
      <Progress class="spinner" bind:loading type="circular" />
      <IconButton class="material-icons" on:click={() => deleteSelf()}
        >done</IconButton
      >
    </div>
  </div>
  <p>{description}</p>
  <div id="bottom">
    <p>{dueDate}</p>
    <p class="text-hint">{createdDate}</p>
    <p class="text-hint">{priorities[priority]}</p>
  </div>
</div>

<style lang="scss">
  .root {
    background-color: var(--clr-height-1-3);
    border-radius: 0.3rem;
    border-left: 0.3rem solid var(--clr-priority);
    padding: 0.7rem 1rem;
    transition-property: transform, height, margin-bottom, padding, opacity;;
    transition-duration: 0.3s;
    margin-bottom: 1rem;
    
    &.deleted {
      transform: translateX(-110%);
      margin-bottom: 0;
      padding: 0 1rem;
    }
  }
  h6 {
    margin: 0;
  }
  #top {
    display: flex;
    justify-content: space-between;
  }
  #bottom {
    display: flex;
  }
  #buttons {
    display: flex;
    align-items: center;
    gap: 1rem;
  }
</style>
