<script lang="ts">
    import Button from "@smui/button";
    import { createEventDispatcher } from "svelte";

    const dispatch = createEventDispatcher();

    // Specifies which main icon is displayed
    export let icon = "code";

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
        mode.
        <br />
        To use the
        <span class="warning__text__code-active">{active}</span>
        mode, a change is required.
        <br />
        <span class="text-disabled">
            This will update the schedule's target mode.
        </span>
    </div>
    <div class="warning__reset-button">
        <Button on:click={() => dispatch("reset")} variant="outlined"
            >Change Mode</Button
        >
    </div>
</div>

<style lang="scss">
    @use "../../../mixins" as *;

    .warning {
        display: flex;
        flex-direction: column;
        justify-content: flex-start;
        align-items: center;
        gap: 1rem;
        margin-top: 4rem;

        @include mobile {
            margin-top: 2rem;
        }

        &__icon {
            display: block;
            color: var(--clr-text-disabled);
            font-size: 5rem;
        }

        &__text {
            max-width: 50%;

            @include widescreen {
                max-width: 60%;
            }

            @include mobile {
                max-width: 100%;
            }

            &__code-active {
                color: var(--clr-primary);
            }
        }
        &__reset-button {
            margin-top: 1rem;
        }
    }
</style>
