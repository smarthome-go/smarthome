<script lang="ts">
    import type { Schedule } from "./types";

    export let data: Schedule;
</script>

<div class="schedule mdc-elevation--z4" >
    {data.data.name}
</div>

<style lang="scss">
</style>
