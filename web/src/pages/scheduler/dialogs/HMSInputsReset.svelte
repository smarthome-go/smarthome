<script lang="ts">
    import Button from "@smui/button";
    import { createEventDispatcher } from "svelte";

    const dispatch = createEventDispatcher();

    // Specifies which main icon is displayed
    export let icon: string = "code";

    // Bound to the `HMSInputs` active mode
    export let active: "hms" | "switches" | "code" = "hms";

    // Bound to the active mode inside the schedule's HMS-code
    export let activeInCode: "hms" | "switches" | "code" = "hms";
</script>

<div class="warning">
    <i class="warning__icon material-icons">{icon}</i>
    <div class="warning__text">
        The last edit was made using the
        <span class="warning__text__code-active">{activeInCode}</span>
        mode. To use the
        <span class="warning__text__code-active">{active}</span>
        mode, a reset is required. However, a reset will cause you loosing your all
        your previous changes.
    </div>
    <div class="warning__reset-button">
        <Button on:click={() => dispatch("reset")} variant="outlined"
            >Reset Code</Button
        >
    </div>
</div>

<style lang="scss">
    .warning {
        display: flex;
        flex-direction: column;
        justify-content: flex-start;
        align-items: center;
        gap: 1rem;
        margin-top: 2rem;

        &__icon {
            display: block;
            color: var(--clr-text-disabled);
            font-size: 5rem;
        }

        &__text {
            max-width: 60%;
            &__code-active {
                color: var(--clr-primary);
            }
        }
        &__reset-button {
            margin-top: 1rem;
        }
    }
</style>
