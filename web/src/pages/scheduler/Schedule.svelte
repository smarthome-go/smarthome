<script lang="ts">
    import IconButton from "@smui/icon-button/src/IconButton.svelte";
    import type { Schedule } from "./main";

    export let data: Schedule;

    // Generates a 12h string from 24h time data
    let timeString = "";
    $: timeString =
        `${data.data.hour <= 12 ? data.data.hour : data.data.hour - 12}`.padStart(2, "0") +
        ":" +
        `${data.data.minute}`.padStart(2, "0") +
        ` ${data.data.hour < 12 ? "AM" : "PM"}`;
</script>

<div class="schedule">
    <span class="schedule__name">{data.data.name}</span>
    <span class="schedule__time">At {timeString}</span>
    <div class="schedule__buttons">
        <IconButton class="material-icons">edit</IconButton>
        <IconButton class="material-icons">cancel</IconButton>
    </div>
</div>

<style lang="scss">
    .schedule {
        height: 5.5rem;
        width: 13rem;
        border-radius: 0.3rem;
        padding: 1rem;
        display: flex;
        flex-direction: column;
        justify-content: space-between;
        background-color: var(--clr-height-1-3);



        &__buttons {
            margin-left: auto;
            display: flex;
            align-items: center;
        }
    }
</style>
