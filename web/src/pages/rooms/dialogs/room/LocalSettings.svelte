<script lang="ts">
    import Button,{ Label } from '@smui/button'
    import Dialog,{ Actions,Content,Title } from '@smui/dialog'
    import FormField from '@smui/form-field'
    import Switch from '@smui/switch'

    export let open = false

    export let powerCamReloadEnabled = false
    export let periodicCamReloadEnabled = false
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content">
    <Title id="title">Local Settings</Title>
    <Content id="content">
        <div id="container">
            <div>
                <span style="color: var(--clr-text);">Description for below</span>
                <FormField>
                    <Switch
                        bind:checked={powerCamReloadEnabled}
                        icons={false}
                    />
                    <span slot="label" class="text-hint"
                        >Switch reload {powerCamReloadEnabled
                            ? 'enabled'
                            : 'disabled'}</span
                    >
                </FormField>
            </div>
            <div>
                <span style="color: var(--clr-text);">Description for below</span>
                <FormField>
                    <Switch
                        bind:checked={periodicCamReloadEnabled}
                        icons={false}
                    />
                    <span slot="label" class="text-hint"
                        >Periodic reload {periodicCamReloadEnabled
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
        gap: 1rem;

        div {
            background-color: var(--clr-height-0-1);
            padding: 1rem;
            border-radius: .4rem;
            display: flex;
            flex-direction: column;
        }
    }
</style>
