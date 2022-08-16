<script lang="ts">
    import Box from "./Box.svelte";
    import { createSnackbar } from "../../global";
    import { onMount } from "svelte";

    const iconMap = {
        "clear sky": ["clear_day", "clear_night"],
        "few clouds": ["partly_cloudy_day", "partly_cloudy_night"],
        "scattered clouds": ["cloud", "cloud"],
        "broken clouds": ["filter_drama", "filter_drama"],
        "shower rain": ["cloudy_snowing", "cloudy_snowing"],
        rain: ["rain", "rain"],
        thunderstorm: ["thunderstorm", "thunderstorm"],
        snow: ["ac_unit", "ac_unit"],
        unknown: ["pending", "pending"],
    };

    let loading = false;
    let loaded = false;

    let data: weatherData = {
        id: 0,
        time: 0,
        weatherTitle: "",
        weatherDescription: "unknown",
        temperature: 0,
        feelsLike: 0,
        humidity: 0,
    };

    interface weatherData {
        id: number;
        time: number;
        weatherTitle: string;
        weatherDescription: string;
        temperature: number;
        feelsLike: number;
        humidity: number;
    }

    // Fetches a current (max. 5 minutes old) weather from the server
    async function loadWeatherData() {
        loading = true;
        try {
            let res = await (await fetch("/api/weather")).json();
            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            // Signal that the weather has been loaded successfully
            data = res;
            loaded = true;
        } catch (err) {
            $createSnackbar(`Could not load weather: ${err}`);
        }
        loading = false;
    }

    onMount(loadWeatherData);
</script>

<Box bind:loading>
    <span slot="header">Weather</span>
    <div class="weather" slot="content">
        <div class="weather__top">
            <i class="material-icons weather__top__icon"
                >{iconMap[data.weatherDescription][0]}</i
            >
            <div class="weather__top__labels">
                <div class="weather__top__labels__title">
                    {data.weatherTitle}
                </div>
                <div class="weather__top__labels__description">
                    {data.weatherDescription}
                </div>
            </div>
        </div>
        <div class="weather__measurements">
            <div class="weather__measurements__temperature">
                {data.temperature} °C
            </div>
            <div class="weather__measurements__feels-like">
                {data.feelsLike} °C
            </div>
            <div class="weather__measurements__humidity">
                {data.humidity}
            </div>
        </div>
    </div>
</Box>

<style lang="scss">
    .weather {
        &__top {
            display: flex;
            align-items: center;
            gap: 1rem;

            &__icon {
                font-size: 4rem;
            }

            &__labels {
                &__title {
                    font-size: 1.125rem;
                    font-weight: bold;
                }
                &__description {
                    font-size: 0.9rem;
                    color: var(--clr-text-hint);
                }
            }
        }
        &__measurements {
            display: flex;
        }
    }
</style>
