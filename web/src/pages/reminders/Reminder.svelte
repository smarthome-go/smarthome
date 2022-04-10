<script lang="ts">
  import Checkbox from "@smui/checkbox";
  import FormField from "@smui/form-field";
  import { onMount } from "svelte";

  export let id: number;
  export let name: string;
  export let description: string;
  export let priority: number;
  export let dueDate: string;
  export let createdDate: string;
  export let userWasNotified: boolean;

  let doneChecked = false;

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
</script>

<div id="main" class="mdc-elevation--z3" style:--clr-priority={priorityColor}>
  <div id="top">
    <h6>{name}</h6>
    <FormField align="end">
      <Checkbox bind:checked={doneChecked} />
      <span slot="label">Mark as completed</span>
    </FormField>
  </div>
  <p>{description}</p>
  <div id="bottom">
    <p>{dueDate}</p>
    <p class="text-hint">{createdDate}</p>
    <p class="text-hint">{priorities[priority]}</p>
  </div>
</div>

<style lang="scss">
  #main {
    background-color: var(--clr-height-1-3);
    border-radius: 0.3rem;
    border-left: 0.3rem solid var(--clr-priority);
    padding: 0.7rem 1rem;
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
</style>
