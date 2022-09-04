<script lang="ts">
    import { createEventDispatcher, onMount } from "svelte";

    import type { Schedule } from "./shared";
    import { timeUntilExecutionText } from "./shared";

    const dispatch = createEventDispatcher();

    export let data: Schedule;

    // Generates a 12h string from 24h time data
    let timeString = "";
    $: timeString =
        `${
            data.data.hour <= 12 ? data.data.hour : data.data.hour - 12
        }`.padStart(2, "0") +
        ":" +
        `${data.data.minute}`.padStart(2, "0") +
        ` ${data.data.hour < 12 ? "AM" : "PM"}`;

    let timeRunning = false;
    let timeUntilString = "";

    // Recursive function which updates the `timeUntilString` every 2000ms
    function updateTimeUntilExecutionText() {
        let now = new Date();
        timeUntilString = timeUntilExecutionText(
            now,
            data.data.hour,
            data.data.minute
        );
        timeRunning =
            data.data.hour === now.getHours() &&
            data.data.minute === now.getMinutes();

        if (timeRunning) {
            // If the schedule is assumed to be executing, hide it after 5 seconds
            setTimeout(() => dispatch("hide"), 5000);
            return;
        }

        setTimeout(updateTimeUntilExecutionText, 1000);
    }

    // Start the time updater
    onMount(updateTimeUntilExecutionText);
</script>

<div class="schedule mdc-elevation--z4">
    <span class="schedule__name">
        {data.data.name}
    </span>
    <div class="schedule__time">
        <span>{timeString}</span>
        <span class="schedule__time__until"
            >{timeRunning ? "Executing" : timeUntilString}</span
        >
    </div>
</div>

<style lang="scss">
    @use "../../../mixins" as *;
    .schedule {
        background-color: var(--clr-height-2-3);
        border-radius: 0.25rem;
        padding: 0.75rem 1rem;
        display: flex;
        flex-direction: column;

        &__name {
            font-weight: bold;
            font-size: 0.9rem;
        }

        &__time {
            font-size: 0.75rem;
            display: flex;
            justify-content: space-between;
            color: var(--clr-text-hint);

            @include mobile {
                flex-direction: column;
            }

            &__until {
                font-size: 0.7rem;
            }
        }
    }
</style>
