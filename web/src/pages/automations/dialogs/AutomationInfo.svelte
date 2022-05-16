<script lang="ts">
    import Button,{ Label } from '@smui/button'
    import Dialog,{
    Actions,
    Content,InitialFocus,
    Title
    } from '@smui/dialog'
    import type { automation } from '../main'

    export let open = false

    export let data: automation
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content">
    <Title id="title">Automation Info</Title>
    <Content id="content" />
    <Actions>
        <Button use={[InitialFocus]}>
            <Label>Close</Label>
        </Button>
    </Actions>
</Dialog>

<style lang="scss">
    @use '../../../mixins' as *;
</style>
