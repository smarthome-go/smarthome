<script lang="ts">
  import { onMount } from "svelte";

  export let id: number;
  export let name: string;
  export let description: string;
  export let priority: number;
  export let dueDate: string;
  export let createdDate: string;
  export let userWasNotified: boolean;

  let priorityColor: string;

  onMount(() => {
    switch (priority) {
      case 0:
        priorityColor = "#ff0000";
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
</script>

<div id="main" class="mdc-elevation--z3" style:--clr-priority={priorityColor}>
  <h5>{name}</h5>
  <p>{description}</p>
  <p>{dueDate}</p>
  <p class="text-hint">{createdDate}</p>
</div>

<style lang="scss">
  #main {
    background-color: var(--clr-height-1-3);
    display: flex;
    border-radius: 0.3rem;
    align-items: center;
  }
</style>
