<script lang="ts">
    import LinearProgress from '@smui/linear-progress'
    import CircularProgress from '@smui/circular-progress'

    export let type: 'linear' | 'circular' = 'linear'
    export let loading = true
    export let id = ''
    let className = ''
    export { className as class }

    let element = type === 'linear' ? LinearProgress : CircularProgress
</script>

<svelte:component this={element} {id} class={
    'custom-progress'
        + (loading ? ' visible' : '')
        + (className !== undefined ? ' ' + className : '')
} indeterminate />

<style lang="scss">
    :global .mdc-circular-progress.custom-progress,
    :global .mdc-linear-progress.custom-progress {
        opacity: 0;
        &.visible { opacity: 1; }
    }
</style>
