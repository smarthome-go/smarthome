<script lang="ts">
    import type { automation } from "./types";

    export let data: automation;
</script>

<div class="automation mdc-elevation--z3">
    <span class="automation__name">
        {data.name}
    </span>
    <span class="automation__description">
        {data.description}
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
        }

        &__description {
            font-size: 0.75rem;
        }
    }
</style>
