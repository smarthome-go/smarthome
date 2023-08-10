<script lang="ts">
    import Button, { Label } from '@smui/button'

    import Dialog, { Actions, Content, Header, InitialFocus, Title } from '@smui/dialog'

    import { createSnackbar } from '../../../global'

    export let open = false

    async function purgeCache() {
        try {
            // Weather cache
            let res = await (
                await fetch('/api/weather/cache', {
                    method: 'DELETE',
                })
            ).json()
            if (!res.success) throw Error(res.error)
            // Power data cache
            res = await (
                await fetch('/api/power/cache', {
                    method: 'DELETE',
                })
            ).json()
            if (!res.success) throw Error(res.error)
        } catch (err) {
            $createSnackbar(`Failed to flush cache: ${err}`)
        }
    }
</script>

<Dialog
    bind:open
    slot="over"
    aria-labelledby="confirm-title"
    aria-describedby="confirm-description"
>
    <Header>
        <Title id="confirm-title">Purge System Cache</Title>
    </Header>
    <Content id="confirm-description"
        >This will delete all non-mandatory data from the system.
        <ul>
            <li>Weather data</li>
            <li>Power usage data</li>
            <li>Homescript URL cache</li>
        </ul>
    </Content>
    <Actions>
        <Button>
            <Label on:click={purgeCache}>Purge</Label>
        </Button>
        <Button defaultAction use={[InitialFocus]}>
            <Label>Cancel</Label>
        </Button>
    </Actions>
</Dialog>
