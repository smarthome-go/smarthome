<script lang="ts">
    import { levels, LogEvent } from "../main";

    export let data: LogEvent = {
        id: 0,
        name: "",
        description: "",
        level: 0,
        time: 0,
    };
</script>

<div class="event" style:--clr-level={levels[data.level].color}>
    <span class="event__name">{data.name}</span>
    <br />
    <span class="event__description text-hint">{data.description}</span>
    <br />
    <span class="event__time text-hint"
        >{new Date(data.time).toUTCString()}</span
    >
</div>

<style lang="scss">
    .event {
        background-color: var(--clr-height-1-3);
        border-radius: 0.3rem;
        padding: 0.5rem 1rem;
        border-left: 0.1rem solid var(--clr-level);

        &__name {
            font-weight: bold;
            font-size: 0.85rem;
        }

        &__description {
            font-size: 0.75rem;
        }

        &__time {
            font-size: 0.7rem;
            font-family: "Jetbrains Mono", monospace;
        }
    }
</style>
