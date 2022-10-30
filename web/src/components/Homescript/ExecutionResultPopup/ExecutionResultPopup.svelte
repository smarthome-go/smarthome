<script lang="ts">
    import Dialog, { Actions, Content, Header, Title } from '@smui/dialog'
    import Button, { Label } from '@smui/button'
    import type { homescriptResponseWrapper } from '../../../homescript'
    import Terminal from './Terminal.svelte'
    import { createEventDispatcher } from 'svelte'

    const dispatch = createEventDispatcher()

    // Keeps track of whether the dialog should be open or not
    export let open = false

    // Data is bound to display the result
    export let data: homescriptResponseWrapper
</script>

<Dialog
    bind:open
    aria-labelledby="title"
    aria-describedby="content"
    fullscreen
    on:SMUIDialog:closed={() => dispatch('close', null)}
>
    <Header>
        <Title id="title">Result of {data.response.id}</Title>
    </Header>
    <Content id="content">
        <div class="status mdc-elevation-z1">
            <h6>Summary</h6>
            <div class="status__container">
                <div class="status__group">
                    <div
                        class="status__indicator mdc-elevation-z3"
                        class:failure={!data.response.success}
                    >
                        <i class="material-icons">{data.response.success ? 'check' : 'error'}</i>
                        {#if !data.modeRun}
                            {data.response.success ? 'Check successful' : 'Errors detected'}
                        {:else}
                            {data.response.success ? 'Run successful' : 'Run failed'}
                        {/if}
                    </div>
                </div>
                <div class="status__group">
                    {#if !data.response.success && data.response.errors.length > 0}
                        <div class="status__error">
                            <i class="material-icons">
                                {#if data.response.errors[0].kind === 'SyntaxError'}
                                    code
                                {:else if data.response.errors[0].kind === 'TypeError'}
                                    tag
                                {:else if data.response.errors[0].kind === 'ReferenceError'}
                                    alt_route
                                {:else if data.response.errors[0].kind === 'ValueError'}
                                    rule
                                {:else if data.response.errors[0].kind === 'RuntimeError' || data.response.errors[0].kind === 'StackOverflow' || data.response.errors[0].kind === 'OutOfBoundsError'}
                                    running_with_errors
                                {:else if data.response.errors[0].kind === 'ThrowError'}
                                    sms_failed
                                {:else}
                                    error
                                {/if}
                            </i>
                            <code>
                                {data.response.errors[0].kind}
                            </code>
                        </div>
                    {/if}
                </div>
            </div>
        </div>
        <div class="output mdc-elevation-z1">
            <h6>Output</h6>
            <Terminal
                data={{
                    code: data.code,
                    modeRun: data.modeRun,
                    exitCode: data.response.exitCode,
                    errors: data.response.errors
                }}
                output={data.response.output}
            />
        </div>
    </Content>
    <Actions>
        <Button
            on:click={() => {
                dispatch('close', null)
            }}
        >
            <Label>Close</Label>
        </Button>
    </Actions>
</Dialog>

<style lang="scss">
    @use '../../../mixins' as *;

    .status {
        background-color: var(--clr-height-0-1);
        padding: 1rem 1.5rem;
        border-radius: 0.3rem;

        &__container {
            margin-top: 1rem;
            display: flex;
            align-items: center;
            justify-content: space-between;

            @include mobile {
                flex-wrap: wrap;
                gap: 1rem;
            }
        }

        &__group {
            display: flex;
            align-items: center;
            gap: 1rem;
        }

        // Summarizes the script's outcome
        &__indicator {
            border: 0.01rem solid var(--clr-success);
            color: var(--clr-success);
            border-radius: 0.4rem;
            padding: 0.3rem 0.6rem;
            display: flex;
            align-items: center;
            gap: 0.6rem;

            i {
                opacity: 80%;
                font-size: 1.2rem;
            }

            &.failure {
                border-color: var(--clr-error);
                color: var(--clr-error);
            }
        }

        // Is shown when the script returned an error
        &__error {
            border: 0.01rem solid var(--clr-error);
            color: var(--clr-error);
            border-radius: 0.4rem;
            padding: 0.3rem 0.6rem;
            display: flex;
            align-items: center;
            gap: 0.6rem;

            code {
                color: var(--clr-error);
            }

            i {
                opacity: 80%;
                font-size: 1.2rem;
            }
        }
    }

    .output {
        background-color: var(--clr-height-0-1);
        padding: 1rem 1.5rem;
        margin-top: 1rem;
        border-radius: 0.3rem;
    }

    h6 {
        margin: 0;
    }
</style>
