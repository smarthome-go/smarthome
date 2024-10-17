<script lang='ts'>
    import HmsFileExplorer from "./HmsFileExplorer.svelte";
    import type { EditorHms } from "./types";
    import type { hmsResWrapper } from "./websocket";


    export let currentData: EditorHms | undefined = undefined
    export let homescripts: EditorHms[] = []
    export let currentExecRes: hmsResWrapper | undefined = undefined
    export let disabled = false
</script>


<div class="files">
    <span class="text-hint mdc-elevation--z2 files__title">Files</span>
    <HmsFileExplorer
        {disabled}
        bind:homescripts
        bind:currentScript={currentData}
    ></HmsFileExplorer>
</div>

<div class="diagnostics">
    <div class="diagnostics__list">
        {#if currentExecRes !== undefined}
            <div class="diagnostics__list__item">
                <span class='icon-info'></span>
                <span>
                {currentExecRes.errors.map(e => e.diagnosticError !== null ? (e.diagnosticError.kind === 1 ? 1 : 0) : 0).reduce((acc, i) => acc + i, 0)}
                </span>
            </div>
            <div class="diagnostics__list__item">
                <span class='icon-warn'></span>
                <span>
                {currentExecRes.errors.map(e => e.diagnosticError !== null ? (e.diagnosticError.kind === 2 ? 1 : 0) : 0).reduce((acc, i) => acc + i, 0)}
                </span>
            </div>
            <div class="diagnostics__list__item">
                <span class='icon-error'></span>
                <span>
                {currentExecRes.errors.map(e => e.diagnosticError !== null ? (e.diagnosticError.kind === 3 ? 1 : 0) : (e.syntaxError !== null ? 1 : 0)).reduce((acc, i) => acc + i, 0)}
                </span>
            </div>
        {:else}
            <span class='text-disabled'>
                No diagnostics available
            </span>
        {/if}
    </div>
</div>

<style lang='scss'>
    @use '../../components/Homescript/HmsEditor/icons.scss' as *;

    .files {
        width: auto;
        display: flex;
        flex-direction: column;

        &__title {
            padding: .3rem .8rem;
        }
    }

    .diagnostics {
        margin-top: auto;

        &__list {
            display: flex;
            gap: .6rem;
            background-color: var(--clr-height-1-4);
            padding: .2rem .5rem;

            &__item {
                display: flex;
                align-items: center;
                gap: .2rem;
                color: var(--clr-text-hint);

                .icon-info {
                    content: url($info-icon-svg);
                    height: 1rem;
                }

                .icon-warn {
                    content: url($warn-icon-svg);
                    height: 1rem;
                }

                .icon-error {
                    content: url($error-icon-svg);
                    height: 1rem;
                }
            }
        }
    }
</style>
