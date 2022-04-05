<script lang="ts">
    export let label: string
    export let uri: string
    export let icon: string

    export let active = false
</script>

{#if active}
    <div class:active>
        <i class="material-icons text-normal">{icon}</i>
        <span class="text-normal">{label}</span>
    </div>
{:else}
    <a class:active href={uri}>
        <i class="material-icons text-normal">{icon}</i>
        <span class="text-normal">{label}</span>
    </a>
{/if}

<style lang="scss">
    a, div {
        height: 3.125rem;
        display: flex;
        align-items: center;
        gap: 0.3rem;
        text-decoration: none;
        color: var(--clr-text);
        margin-top: 0.625rem;
        border-radius: 0.4rem;
        transition: background-color 0.2s;

        &:hover, &.active { background-color: var(--clr-hover); }
    }

    i {
        height: 100%;
        aspect-ratio: 1;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: var(--icon-size);
    }
</style>
