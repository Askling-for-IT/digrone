import Navigation from '@/components/Navigation';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { useQuery } from '@tanstack/react-query';
import { fetchMatches, fetchPlayerStats, fetchBeerBalances, type Match } from '@/lib/api';

type CompletedMatch = Match & {
  goals_for: number;
  goals_against: number;
};

const podiumColors = [
  'bg-yellow-400 text-yellow-900 dark:bg-yellow-700 dark:text-yellow-100',
  'bg-slate-300 text-slate-900 dark:bg-slate-700 dark:text-slate-100',
  'bg-orange-400 text-orange-900 dark:bg-orange-700 dark:text-orange-100',
];

const Statistics = () => {
  const query = useQuery({
    queryKey: ['statistics'],
    queryFn: async () => {
      const [matches, playerStats, beerBalances] = await Promise.all([
        fetchMatches(),
        fetchPlayerStats(),
        fetchBeerBalances(),
      ]);
      return { matches, playerStats, beerBalances };
    },
    staleTime: 60_000,
  });

  if (query.isPending) {
    return <p>Laddar statistik...</p>;
  }

  if (query.isError) {
    return <p role="alert">Fel vid hämtning av statistik: {query.error.message || 'Okänt fel'}</p>;
  }

  const { matches, playerStats: players, beerBalances } = query.data;
  const sortedByGoals = [...players].sort(
    (a, b) => b.goals - a.goals || b.points - a.points || a.name.localeCompare(b.name, 'sv'),
  );
  const sortedByPoints = [...players].sort(
    (a, b) => b.points - a.points || b.goals - a.goals || a.name.localeCompare(b.name, 'sv'),
  );
  const beersWithHistory = beerBalances.filter(
    (player) => player.fined_beers > 0 || player.paid_beers > 0,
  );
  const completedMatches = matches
    .filter(
      (match): match is CompletedMatch =>
        match.status === 'completed' &&
        match.goals_for !== null &&
        match.goals_against !== null,
    )
    .sort((a, b) => Date.parse(a.start_time) - Date.parse(b.start_time));

  return (
    <div className="min-h-screen bg-background">
      <Navigation />
      <main className="container mx-auto px-4 py-6 md:py-12">
        <h1 className="mb-6 text-3xl font-bold text-primary md:mb-8 md:text-4xl">Statistik</h1>
        <p className="text-xl font-bold mb-4">Serie + turneringar</p>
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-8">
          {/* Podium for Top 3 Goal Scorers */}
          {players.length >= 3 && (
            <div>
              <h2 className="text-2xl font-bold text-center mb-4">Mål</h2>

              {/* Mobile podium */}
              <ol className="space-y-2 md:hidden">
                {sortedByGoals
                  .slice(0, 3)
                  .map((player, index) => (
                    <li 
                      key={player.player_id}
                      className={`flex items-center gap-3 p-3 rounded-lg ${podiumColors[index]}`}
                    >
                      <span aria-hidden="true">
                        {['🥇', '🥈', '🥉'][index]}
                      </span>
                      <span className="min-w-0 flex-1 break-words">
                        {player.name}
                      </span>
                      <span className="shrink-0 font-bold">
                        {player.goals} mål
                      </span>
                    </li>
                  ))
                }
              </ol>

              {/* Desktop podium */}
              <div className="hidden items-end justify-center gap-4 py-8 md:flex">
                <>
                  {/* 2nd place */}
                  <div className="flex flex-col items-center bg-slate-300 dark:bg-slate-700 p-6 rounded-lg w-40" style={{height: '200px'}}>
                    <div className="text-4xl mb-2">🥈</div>
                    <div className="font-bold text-center text-slate-900 dark:text-slate-100">{sortedByGoals[1].name}</div>
                    <div className="text-3xl font-bold text-slate-900 dark:text-slate-100 mt-auto">{sortedByGoals[1].goals}</div>
                    <div className="text-xs text-slate-700 dark:text-slate-300">mål</div>
                  </div>
                  {/* 1st place - tallest */}
                  <div className="flex flex-col items-center bg-yellow-400 dark:bg-yellow-700 p-6 rounded-lg w-40" style={{height: '240px'}}>
                    <div className="text-5xl mb-2">🥇</div>
                    <div className="font-bold text-lg text-center text-yellow-900 dark:text-yellow-100">{sortedByGoals[0].name}</div>
                    <div className="text-4xl font-bold text-yellow-900 dark:text-yellow-100 mt-auto">{sortedByGoals[0].goals}</div>
                    <div className="text-sm text-yellow-800 dark:text-yellow-200">mål</div>
                  </div>
                  {/* 3rd place */}
                  <div className="flex flex-col items-center bg-orange-400 dark:bg-orange-700 p-4 rounded-lg w-40" style={{height: '165px'}}>
                    <div className="text-3xl mb-2">🥉</div>
                    <div className="font-bold text-center text-orange-900 dark:text-orange-100">{sortedByGoals[2].name}</div>
                    <div className="text-2xl font-bold text-orange-900 dark:text-orange-100 mt-auto">{sortedByGoals[2].goals}</div>
                    <div className="text-xs text-orange-800 dark:text-orange-200">mål</div>
                  </div>
                </>
              </div>
            </div>
          )}

          {/* Podium for Top 3 Points Leaders */}
          {players.length >= 3 && (
            <div>
              <h2 className="text-2xl font-bold text-center mb-4">Totalpoäng</h2>

              {/* Mobile podium */}
              <ol className="space-y-2 md:hidden">
                {sortedByPoints
                  .slice(0, 3)
                  .map((player, index) => (
                    <li 
                      key={player.player_id}
                      className={`flex items-center gap-3 p-3 rounded-lg ${podiumColors[index]}`}
                    >
                      <span aria-hidden="true">
                        {['🥇', '🥈', '🥉'][index]}
                      </span>
                      <span className="min-w-0 flex-1 break-words">
                        {player.name}
                      </span>
                      <span className="shrink-0 opacity-75 text-sm">
                        ({player.goals}+{player.assists})
                      </span>
                      <span className="shrink-0 font-bold">
                        {player.points} poäng
                      </span>
                    </li>
                  ))
                }
              </ol>

              {/* Desktop podium */}
              <div className="hidden items-end justify-center gap-4 py-8 md:flex">
                <>
                  {/* 2nd place */}
                  <div className="flex flex-col items-center bg-slate-300 dark:bg-slate-700 p-6 rounded-lg w-40" style={{height: '200px'}}>
                    <div className="text-4xl mb-2">🥈</div>
                    <div className="font-bold text-center text-slate-900 dark:text-slate-100">{sortedByPoints[1].name}</div>
                    <div className="text-3xl font-bold text-slate-900 dark:text-slate-100 mt-auto">{sortedByPoints[1].points}</div>
                    <div className="text-xs text-slate-700 dark:text-slate-300">{sortedByPoints[1].goals}+{sortedByPoints[1].assists}</div>
                  </div>
                  {/* 1st place - tallest */}
                  <div className="flex flex-col items-center bg-yellow-400 dark:bg-yellow-700 p-6 rounded-lg w-40" style={{height: '240px'}}>
                    <div className="text-5xl mb-2">🥇</div>
                    <div className="font-bold text-lg text-center text-yellow-900 dark:text-yellow-100">{sortedByPoints[0].name}</div>
                    <div className="text-4xl font-bold text-yellow-900 dark:text-yellow-100 mt-auto">{sortedByPoints[0].points}</div>
                    <div className="text-sm text-yellow-800 dark:text-yellow-200">{sortedByPoints[0].goals}+{sortedByPoints[0].assists}</div>
                  </div>
                  {/* 3rd place */}
                  <div className="flex flex-col items-center bg-orange-400 dark:bg-orange-700 p-4 rounded-lg w-40" style={{height: '165px'}}>
                    <div className="text-3xl mb-2">🥉</div>
                    <div className="font-bold text-center text-orange-900 dark:text-orange-100">{sortedByPoints[2].name}</div>
                    <div className="text-2xl font-bold text-orange-900 dark:text-orange-100 mt-auto">{sortedByPoints[2].points}</div>
                    <div className="text-xs text-orange-800 dark:text-orange-200">{sortedByPoints[2].goals}+{sortedByPoints[2].assists}</div>
                  </div>
                </>
              </div>
            </div>
          )}
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-8">
          <Card>
            <CardHeader>
              <CardTitle>Böter</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="max-h-[600px] overflow-y-auto">
                <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Spelare</TableHead>
                    <TableHead className="text-right">Aktuell skuld</TableHead>
                    <TableHead className="text-right">Betald skuld</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {beersWithHistory.length > 0 ? (
                    beersWithHistory.map((player) => (
                        <TableRow key={player.player_id}>
                          <TableCell>{player.name}</TableCell>
                          <TableCell className="text-right font-semibold">
                            {player.outstanding_beers < 0
                              ? `${Math.abs(player.outstanding_beers)} burkar tillgodo`
                              : player.outstanding_beers > 0
                              ? `${player.outstanding_beers} burkar`
                              : '-'}
                          </TableCell>
                          <TableCell className="text-right text-muted-foreground">
                            {player.paid_beers > 0 ? `${player.paid_beers} burkar` : '-'}
                          </TableCell>
                        </TableRow>
                      ))
                  ) : (
                    <TableRow>
                      <TableCell colSpan={3} className="text-center text-muted-foreground">
                        Ingen bötesinformation tillgänglig ännu
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Spelade matcher</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="max-h-[600px] overflow-y-auto">
                <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Datum</TableHead>
                    <TableHead>Motståndare</TableHead>
                    <TableHead className="text-center">Resultat</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {completedMatches.length > 0 ? (
                    completedMatches.map((match) => {
                      const won = match.goals_for > match.goals_against;
                      const draw = match.goals_for === match.goals_against;
                      return (
                        <TableRow key={match.id}>
                          <TableCell>{new Date(match.start_time).toLocaleDateString(
                            'sv-SE', { 
                              day: 'numeric',
                              month: 'numeric',
                              year: 'numeric',
                              hour: '2-digit',
                              minute: '2-digit'
                            })}
                          </TableCell>
                          <TableCell>{match.opponent}</TableCell>
                          <TableCell className="text-center">
                            <span
                              className={`font-semibold ${
                                won
                                  ? 'text-primary'
                                  : draw
                                  ? 'text-muted-foreground'
                                  : 'text-destructive'
                              }`}
                            >
                              {match.goals_for} - {match.goals_against}
                            </span>
                          </TableCell>
                        </TableRow>
                      );
                    })
                  ) : (
                    <TableRow>
                      <TableCell colSpan={3} className="text-center text-muted-foreground">
                        Inga spelade matcher registrerade ännu
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
              </div>
            </CardContent>
          </Card>
        </div>

        <div className="grid gap-8">
          <Card>
            <CardHeader>
              <CardTitle>Skytteliga</CardTitle>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-16">Plac.</TableHead>
                    <TableHead>Spelare</TableHead>
                    <TableHead className="text-right">Poäng</TableHead>
                    <TableHead className="text-right">Mål</TableHead>
                    <TableHead className="text-right">Assist</TableHead>
                    <TableHead className="text-right">Utv. minuter</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {players.length > 0 ? (
                    sortedByPoints.map((player, index) => (
                        <TableRow key={player.player_id}>
                          <TableCell className="font-medium">{index + 1}</TableCell>
                          <TableCell>{player.name}</TableCell>
                          <TableCell className="text-right font-semibold">
                            {player.points}
                          </TableCell>
                          <TableCell className="text-right font-semibold">{player.goals}</TableCell>
                          <TableCell className="text-right">{player.assists}</TableCell>
                          <TableCell className="text-right">{player.penalty_minutes}</TableCell>
                        </TableRow>
                      ))
                  ) : (
                    <TableRow>
                      <TableCell colSpan={6} className="text-center text-muted-foreground">
                        Ingen statistik tillgänglig ännu
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </CardContent>
          </Card>

        </div>
      </main>
    </div>
  );
};

export default Statistics;
