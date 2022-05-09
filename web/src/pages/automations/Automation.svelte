<script lang="ts">
    import type { automation } from './main'

    export let data: automation
</script>

<div class="automation">
    <span class="automation__name">{data.name}</span>
    <span class="automation__description">{data.description}</span>
    <span class="automation__cronDescription">{data.cronDescription}</span>
    <span class="automation__homescriptId">{data.homescriptId}</span>
</div>

<style lang="scss">
    .automation {
        height: 4rem;
        width: 15rem;
        border-radius: .4rem;
        background-color: var(--clr-height-0-4)
    }
</style>
