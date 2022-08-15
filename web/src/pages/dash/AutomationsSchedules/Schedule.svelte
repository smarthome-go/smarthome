<script lang="ts">
    import { createEventDispatcher, onMount } from "svelte";

    import type { Schedule } from "./types";

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

    // Is used to calculate the time until the schedule's execution
    // Returns a user-friendly string
    export function timeUntilExecutionText(
        now: Date,
        hourThen: number,
        minuteThen: number
    ): string {
        now.setTime(now.getTime());
        const minuteNow = now.getMinutes();
        const hourNow = now.getHours();
        let hourDifference = hourThen - hourNow;
        const minuteDifference = minuteThen - minuteNow;
        let outputText = "In ";

        if (minuteDifference < 0) hourDifference--;

        if (hourDifference < 0) hourDifference += 24;

        if (hourDifference > 0) {
            outputText +=
                hourDifference > 1
                    ? `${hourDifference} hours`
                    : `${hourDifference} hour`;
        }

        if (hourDifference !== 0 && minuteDifference !== 0)
            outputText += " and ";

        if (hourDifference === 0 && minuteDifference === 1) {
            outputText += ` ${60 - now.getSeconds()} seconds`;
        } else if (minuteDifference > 0) {
            outputText +=
                minuteDifference > 1
                    ? `${minuteDifference} minutes`
                    : `${minuteDifference} minute`;
        } else if (minuteDifference < 0) {
            outputText +=
                minuteDifference + 60 > 1
                    ? `${minuteDifference + 60} minutes`
                    : `${minuteDifference + 60} minute`;
        }
        return outputText;
    }

    // Recursive function which updates the `timeUntilString` every 2000ms
    function updateTimeUntilExecutionText() {
        timeUntilString = timeUntilExecutionText(
            new Date(),
            data.data.hour,
            data.data.minute
        );
        timeRunning =
            data.data.hour === new Date().getHours() &&
            data.data.minute === new Date().getMinutes();

        // If the schedule is assumed to be executing, hide it after 5 seconds
        setTimeout(() => dispatch("hide"), 5000);

        setTimeout(updateTimeUntilExecutionText, 2000);
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
        <span class="schedule__time__until">{timeUntilString}</span>
    </div>
</div>

<style lang="scss">
    .schedule {
        background-color: var(--clr-height-2-3);
        border-radius: 0.25rem;
        padding: 0.75rem 1rem;
        display: flex;
        flex-direction: column;

        &__name {
            font-weight: bold;
        }

        &__time {
            font-size: 0.9rem;

            &__until {
                color: var(--clr-text-hint);
            }
        }
    }
</style>
