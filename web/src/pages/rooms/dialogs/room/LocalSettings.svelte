<script lang="ts">
    import Button,{ Label } from '@smui/button'
    import Dialog,{ Actions,Content,Title } from '@smui/dialog'
    import FormField from '@smui/form-field'
    import Switch from '@smui/switch'
    import { periodicCamReloadEnabled,powerCamReloadEnabled } from '../../main'

    export let open = false

    function onChange() {
        console.log('change')
        localStorage.setItem(
            'smarthome_periodic_cam_reload_enabled',
            `${$periodicCamReloadEnabled}`
        )
        localStorage.setItem(
            'smarthome_power_cam_reload_enabled',
            `${$powerCamReloadEnabled}`
        )
    }
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content">
    <Title id="title">Local Settings</Title>
    <Content id="content">
        <div id="container">
            <div>
                <span style="color: var(--clr-text);"
                    >Reload cameras on power change</span
                >
                <FormField>
                    <Switch
                        bind:checked={$powerCamReloadEnabled}
                        on:SMUISwitch:change={onChange}
                        icons={false}
                    />
                    <span slot="label" class="text-hint indicator"
                        >Switch reload {$powerCamReloadEnabled
                            ? 'enabled'
                            : 'disabled'}</span
                    >
                </FormField>
            </div>
            <div>
                <span style="color: var(--clr-text);"
                    >Reload cameras every 10 seconds</span
                >
                <FormField>
                    <Switch
                        bind:checked={$periodicCamReloadEnabled}
                        on:SMUISwitch:change={onChange}
                        icons={false}
                    />
                    <span slot="label" class="text-hint indicator"
                        >Periodic reload {$periodicCamReloadEnabled
                            ? 'enabled'
                            : 'disabled'}</span
                    >
                </FormField>
            </div>
        </div>
    </Content>
    <Actions>
        <Button>
            <Label>Close</Label>
        </Button>
    </Actions>
</Dialog>

<style lang="scss">
    #container {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 1rem;

        div {
            background-color: var(--clr-height-0-1);
            padding: 1rem;
            border-radius: 0.4rem;
            display: flex;
            flex-direction: column;
        }
    }
</style>
