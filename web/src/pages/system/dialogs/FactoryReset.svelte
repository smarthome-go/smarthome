<script lang="ts">
    import Button, { Label } from '@smui/button'
    import Checkbox from '@smui/checkbox'
    import Dialog, { Actions, Content, Header, Title } from '@smui/dialog'
    import FormField from '@smui/form-field'
    import Progress from '../../../components/Progress.svelte'
    import { createSnackbar, sleep } from '../../../global'

    export let open = false

    let confirm = false

    let resetStarted = false
    let resetRunning = false

    let statusMessage = ''
    let statusError = ''

    let remainingSecs = 0

    async function redirect() {
        for (remainingSecs = 5; remainingSecs > 0; remainingSecs--) await sleep(1000)
        if (confirm) window.location.href = '/logout'
    }

    async function doReset() {
        resetRunning = true
        resetStarted = true
        try {
            const res = await (
                await fetch('/api/system/config/factory', {
                    method: 'DELETE',
                    headers: { 'Content-Type': 'application/json' },
                })
            ).json()
            if (!res.success) {
                statusMessage = res.message
                statusError = res.error
            } else {
                redirect()
            }
        } catch (err) {
            $createSnackbar(`Could not perform factory reset: ${err}`)
        }
        resetRunning = false
    }
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content">
    <Header>
        <Title id="title">Factory Reset</Title>
    </Header>
    <Content id="content">
        {#if resetStarted}
            {#if statusMessage !== '' || statusError !== ''}
                <span id="reset-output">
                    {statusMessage}
                    <br />
                    <span style="color: var(--clr-error)">
                        {statusError}
                    </span>
                </span>
            {:else if !resetRunning}
                <span style="color: var(--clr-success)">
                    <strong>Success: </strong>
                    this server has been reset to factory settings.
                </span>
                <br />
                <span class="text-hint">
                    You will be redirected to the login screen in {remainingSecs} second(s)
                </span>
            {:else}
                <Progress bind:loading={resetRunning} />
                Performing factory reset...
            {/if}
        {:else}
            <div id="confirm">
                <div class="list warn">
                    <span>Before you continue</span>
                    <ul>
                        <li>
                            This action will <strong>erase all data</strong> on this server
                        </li>
                        <li>
                            Do <strong>not shutdown</strong> the server during the process
                        </li>
                        <li>Proceeding comes at your own risk</li>
                        <li>A factory reset may take some time</li>
                    </ul>
                </div>

                <FormField>
                    <Checkbox id="reset-confirm" bind:checked={confirm} />
                    <span slot="label">I know the risk and want to continue.</span>
                </FormField>
                <Button id="reset-button" variant="raised" disabled={!confirm} on:click={doReset}>
                    <Label>Reset</Label>
                </Button>
            </div>
        {/if}
    </Content>
    <Actions>
        <Button
            on:click={() => {
                confirm = false
            }}
        >
            <Label>Cancel</Label>
        </Button>
    </Actions>
</Dialog>

<style lang="scss">
    @use '../../../mixins' as *;

    #confirm {
        width: 30rem;

        @include mobile {
            width: auto;
        }
    }

    .list {
        margin-bottom: 0.5rem;

        &.warn {
            span {
                color: var(--clr-error);
            }
        }

        span {
            font-weight: bold;
            color: var(--clr-primary);
        }

        ul {
            padding: 0 1rem;
            margin: 0;
            margin-top: 0.125rem;
        }
    }
    :global #reset-confirm {
        --mdc-theme-secondary: var(--clr-error);
        --mdc-ripple-color: var(--clr-error);
    }
    :global #reset-button {
        --mdc-theme-primary: var(--clr-error);
        margin-top: 0.8rem;
        display: block;
    }
    #reset-output {
        font-family: 'Jetbrains Mono', monospace;
        font-size: 0.9rem;
    }
</style>
