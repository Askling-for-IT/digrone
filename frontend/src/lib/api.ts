export type Match = {
    id: string;
    start_time: string;
    opponent: string;
    home_game: boolean;
    competition: string;
    status: 'scheduled' | 'completed' | 'cancelled';
    goals_for: number;
    goals_against: number;
};

export async function fetchMatches(): Promise<Match[]> {
    const response = await fetch('/api/matches');
    if (!response.ok) {
        throw new Error('Failed to fetch matches');
    }
    return response.json();
}