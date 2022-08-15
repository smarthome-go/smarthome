<script lang="ts">
    import type { automationWrapper } from "./types";

    export let data: automationWrapper;

    // Generates a 12h string from 24h time data
    let timeString = "";
    $: timeString =
        `${
            data.hours <= 12 ? data.hours : data.hours - 12
        }`.padStart(2, "0") +
        ":" +
        `${data.minutes}`.padStart(2, "0") +
        ` ${data.hours < 12 ? "AM" : "PM"}`;
</script>

<div class="automation mdc-elevation--z3">
    <span class="automation__name">
        {data.data.name}
    </span>
    <span class="automation__time">
        {timeString}
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
            color: var(--clr-text-hint);
            font-size: 0.75rem;
        }
    }
</style>
