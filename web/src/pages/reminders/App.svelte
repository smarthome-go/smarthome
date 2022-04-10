<script lang="ts">
  import Button from "@smui/button";
  import IconButton from "@smui/icon-button";
  import SegmentedButton,{ Label,Segment } from "@smui/segmented-button";
  import Textfield from "@smui/textfield";
  import CharacterCounter from "@smui/textfield/character-counter";
  import HelperText from "@smui/textfield/helper-text";
  import { onMount } from "svelte";
  import Progress from "../../components/Progress.svelte";
  import { createSnackbar } from "../../global";
  import Page from "../../Page.svelte";
  import { reminder,reminders } from "./main";
  import Reminder from "./Reminder.svelte";

  // Add new inputs
  let inputName = "";
  let inputDescription = "";
  let selectedPriority = "Normal";
  const priorities = ["Low", "Normal", "Medium", "High", "Urgent"];

  let loading = false;

  async function loadReminders() {
    loading = true;
    try {
      const res = (await (
        await fetch("/api/reminder/list")
      ).json()) as reminder[];
      console.log(res);
      reminders.set(res);
    } catch (err) {
      $createSnackbar("Could not load reminders");
    }
    loading = false;
  }

  onMount(() => loadReminders());
</script>

<Page>
  <Progress id="loader" bind:loading />
  <div id="content">
    <div id="container" class="mdc-elevation--z1">
      <div id="header">
        <h6>Reminders</h6>
        <IconButton class="material-icons" on:click={() => loadReminders()}
          >refresh</IconButton
        >
      </div>
      <div class="reminders" class:empty={$reminders.length === 0}>
        {#if $reminders.length === 0}
          No reminders
        {/if}
        {#each $reminders as reminder (reminder.id)}
          <Reminder {...reminder} />
        {/each}
      </div>
    </div>
    <div id="add" class="mdc-elevation--z1">
      <div id="name">
        <Textfield
          style="width: 100%;"
          helperLine$style="width: 100%;"
          bind:value={inputName}
          label="Name"
          input$maxlength={100}
        >
          <CharacterCounter slot="helper">0 / 100</CharacterCounter>
        </Textfield>
      </div>
      <div id="description">
        <Textfield
          style="width: 100%;"
          helperLine$style="width: 100%;"
          textarea
          bind:value={inputDescription}
          label="Description"
          input$rows={5}
        >
          <HelperText slot="helper"
            >Describe which task you want to accomplish</HelperText
          >
        </Textfield>
      </div>
      <SegmentedButton
        segments={priorities}
        let:segment
        singleSelect
        bind:selected={selectedPriority}
      >
        <Segment {segment}>
          <Label>{segment}</Label>
        </Segment>
      </SegmentedButton>
      
      <br />
      <br />
      <Button on:click={() => {}} touch variant="raised">
        <Label>Create</Label>
      </Button>

      <Button on:click={() => {}} touch>
        <Label>Cancel</Label>
      </Button>
    </div>
  </div>
</Page>

<style lang="scss">
  @use "../../mixins" as *;
  #content {
    display: flex;
    flex-direction: column;
    @include widescreen {
      flex-direction: row;
      gap: 2rem;
    }
    margin: 1rem 1.5rem;
    gap: 1rem;
  }
  #container {
    background-color: var(--clr-height-0-1);
    border-radius: 0.4rem;
    padding: 1.5rem;
    @include widescreen {
      width: 50%;
    }
  }
  #add {
    background-color: var(--clr-height-0-1);
    border-radius: 0.4rem;
    padding: 1.5rem;
    @include widescreen {
      width: 50%;
    }
  }
  .reminders {
    padding: 1rem 0;
    display: flex;
    flex-direction: column;
    overflow-x: hidden;

    &.empty {
      display: flex;
      align-items: center;
      justify-content: center;
    }
  }
  #header {
    display: flex;
    justify-content: space-between;
    h6 {
      margin: 0;
    }
  }
  #description {
    margin-top: 1rem;
    :global(.mdc-text-field__resizer) {
      resize: none;
    }
  }
</style>
