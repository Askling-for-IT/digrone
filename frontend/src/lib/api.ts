export type Match = {
    id: string;
    start_time: string;
    opponent: string;
    home_game: boolean;
    competition: string;
    status: 'scheduled' | 'completed' | 'cancelled';
    goals_for: number | null;
    goals_against: number | null;
    statistics_in_historical_totals: boolean;
};

export type PlayerStats = {
    player_id: string;
    name: string;
    goals: number;
    assists: number;
    points: number;
    penalty_minutes: number;
};

export type Beer = {
    player_id: string;
    name: string;
    fined_beers: number;
    paid_beers: number;
    outstanding_beers: number;
};

export async function fetchMatches(): Promise<Match[]> {
    const response = await fetch('/api/matches');
    if (!response.ok) {
        throw new Error('Failed to fetch matches');
    }
    return response.json();
}

export async function fetchPlayerStats(): Promise<PlayerStats[]> {
    const response = await fetch('/api/statistics/players');
    if (!response.ok) {
        throw new Error('Failed to fetch player stats');
    }
    return response.json();
}

export async function fetchBeerBalances(): Promise<Beer[]> {
    const response = await fetch('/api/statistics/beers');
    if (!response.ok) {
        throw new Error('Failed to fetch beer balances');
    }
    return response.json();
}
