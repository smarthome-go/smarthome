<script lang="ts">
    import { createEventDispatcher, onMount } from "svelte";

    import type { automationWrapper } from "./shared";
    import { timeUntilExecutionText } from "./shared";

    const dispatch = createEventDispatcher();

    export let data: automationWrapper;

    // Generates a 12h string from 24h time data
    let timeString = "";
    $: timeString =
        `${data.hours <= 12 ? data.hours : data.hours - 12}`.padStart(2, "0") +
        ":" +
        `${data.minutes}`.padStart(2, "0") +
        ` ${data.hours < 12 ? "AM" : "PM"}`;

    let timeRunning = false;
    let timeUntilString = "";

    // Recursive function which updates the `timeUntilString` every 2000ms
    function updateTimeUntilExecutionText() {
        let now = new Date();
        timeUntilString = timeUntilExecutionText(now, data.hours, data.minutes);

        timeRunning =
            data.hours === now.getHours() && data.minutes === now.getMinutes();

        if (timeRunning) {
            // If the schedule is assumed to be executing, hide it after 5 seconds
            setTimeout(() => dispatch("hide"), 5000);
            return
        }

        setTimeout(updateTimeUntilExecutionText, 1000);
    }

    // Start the time updater
    onMount(updateTimeUntilExecutionText);
</script>

<div class="automation mdc-elevation--z3">
    <span class="automation__name">
        {data.data.name}
    </span>
    <span class="automation__time">
        {timeString}
        <span class="automation__time__until"
            >{timeRunning ? "Executing" : timeUntilString}</span
        >
    </span>
</div>

<style lang="scss">
    .automation {
        background-color: var(--clr-height-2-3);
        border-radius: 0.25rem;
        padding: 0.75rem 1rem;
        display: flex;
        flex-direction: column;

        &__name {
            font-weight: bold;
            font-size: 0.9rem;
            white-space: nowrap;
            text-overflow: ellipsis;
            overflow: hidden;
        }

        &__time {
            font-size: 0.75rem;
            display: flex;
            justify-content: space-between;
            color: var(--clr-text-hint);

            &__until {
                font-size: 0.7rem;
            }
        }
    }
</style>
