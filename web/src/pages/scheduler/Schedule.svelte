<script lang="ts">
    import IconButton from "@smui/icon-button/src/IconButton.svelte";
    import { createEventDispatcher } from "svelte";
    import ConfirmDeletion from "./dialogs/ConfirmDeletion.svelte";
    import EditSchedule from "./dialogs/EditSchedule.svelte";
    import type { Schedule } from "./main";

    export let data: Schedule;

    // Specifies whether the edit dialog should be open or not
    let editOpen: boolean = false;
    // Specifies whether the delete dialog should be open or not
    let deleteOpen: boolean = false;

    // Event dispatcher
    const dispatch = createEventDispatcher();

    // Generates a 12h string from 24h time data
    let timeString = "";
    $: timeString =
        `${
            data.data.hour <= 12 ? data.data.hour : data.data.hour - 12
        }`.padStart(2, "0") +
        ":" +
        `${data.data.minute}`.padStart(2, "0") +
        ` ${data.data.hour < 12 ? "AM" : "PM"}`;
</script>

<EditSchedule bind:data bind:open={editOpen} />
<ConfirmDeletion
    bind:open={deleteOpen}
    name={data.data.name}
    on:confirm={() => {
        dispatch("delete");
    }}
/>

<div class="schedule">
    <span class="schedule__name">{data.data.name}</span>
    <span class="schedule__time">At {timeString}</span>
    <div class="schedule__buttons">
        <IconButton class="material-icons" on:click={() => (editOpen = true)}
            >edit</IconButton
        >
        <IconButton class="material-icons" on:click={() => (deleteOpen = true)}
            >cancel</IconButton
        >
    </div>
</div>

<style lang="scss">
    .schedule {
        height: 5.5rem;
        width: 17rem;
        border-radius: 0.3rem;
        padding: 1rem;
        display: flex;
        flex-direction: column;
        justify-content: space-between;
        background-color: var(--clr-height-1-3);

        &__time {
            font-size: 0.85rem;
        }

        &__buttons {
            margin-left: auto;
            display: flex;
            align-items: center;
        }
    }
</style>
