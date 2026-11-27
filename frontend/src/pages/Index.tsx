import Navigation from '@/components/Navigation';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Trophy, Award, Target } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { fetchMatches, type Match } from '@/lib/api';

type CompletedMatch = Match & { 
  goals_for: number;
  goals_against: number;
};

const byStartTime = (a: Match, b: Match) =>
  Date.parse(a.start_time) - Date.parse(b.start_time);

const Index = () => {
  const query = useQuery({
    queryKey: ['matches'],
    queryFn: fetchMatches,
    staleTime: 60_000,
  });

  if (query.isPending) {
    return (
      <div className="min-h-screen bg-background">
        <Navigation />
        <main className="container mx-auto px-4 py-6" role="status">Laddar matcher...</main>
      </div>
    );
  }

  if (query.isError) {
    return (
      <div className="min-h-screen bg-background">
        <Navigation />
        <main className="container mx-auto px-4 py-6" role="alert">Kunde inte hämta matcher. Försök igen senare.</main>
      </div>
    );
  }

  const matches = query.data;

  const completed = matches.filter(
    (m): m is CompletedMatch => 
      m.status === 'completed' &&
      m.goals_for !== null &&
      m.goals_against !== null,
  ).sort(byStartTime);

  const stats = {
    matches: completed.length,
    wins: completed.filter(
      (m) => m.goals_for > m.goals_against,
    ).length,
    totGoalsFor: completed.reduce(
      (sum, m) => sum + m.goals_for,
      0,
    ),
    totGoalsAgainst: completed.reduce(
      (sum, m) => sum + m.goals_against,
      0,
    ),
  };

  const lastFiveResults = completed
    .slice(-5)
    .map((m) => {
      if (m.goals_for > m.goals_against) return 'win';
      if (m.goals_for < m.goals_against) return 'loss';
      return 'draw';
    });

  const nextMatch = matches
    .filter((m) => 
      m.status === 'scheduled' &&
      Date.parse(m.start_time) > Date.now(),
    )
    .sort(byStartTime)[0] ?? null;

  const prevMatch = completed
    .sort(byStartTime)
    .slice(-1)[0] ?? null;


  return (
    <div className="min-h-screen bg-background">
      <Navigation />
      <main className="container mx-auto px-4">
        <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', margin: '20px' }}>
          <img src="digrone_logo.png" alt="Di Gröne logo"></img>
        </div>
        <div className="grid md:grid-cols-3 gap-6 mb-12">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Matcher spelade</CardTitle>
              <Trophy className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.matches}</div>
              <p className="text-xs text-muted-foreground">{stats.wins} vinster · Alla tävlingar</p>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Senaste 5 matcherna</CardTitle>
              <Award className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="flex gap-2 justify-center items-center h-12">
                {lastFiveResults.length > 0 ? (
                  lastFiveResults.map((result, idx) => (
                    <div
                      key={idx}
                      className={`w-8 h-8 rounded-full ${
                        result === 'win' ? 'bg-green-500' :
                        result === 'loss' ? 'bg-red-500' :
                        'bg-gray-400'
                      }`}
                      title={result === 'win' ? 'Vinst' : result === 'loss' ? 'Förlust' : 'Oavgjort'}
                    />
                  ))
                ) : (
                  <p className="text-xs text-muted-foreground">Inga matcher än</p>
                )}
              </div>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Målratio</CardTitle>
              <Target className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.totGoalsFor} - {stats.totGoalsAgainst}</div>
              <p className="text-xs text-muted-foreground">Gjorda mål - Insläppta mål</p>
            </CardContent>
          </Card>
        </div>

        <div className="grid md:grid-cols-2 gap-6 mb-12">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-0">
              <CardTitle className="font-medium">Föregående match</CardTitle>
            </CardHeader>
            <CardContent>
              {prevMatch ? (
                <div className="mt-4 text-center">
                  <p className="text-sm text-muted-foreground mb-2">
                    {new Date(prevMatch.start_time).toLocaleDateString('sv-SE', { 
                      day: 'numeric',
                      month: 'short',
                      year: 'numeric'
                    })}
                  </p>
                  <p className="text-lg mb-2">
                    {prevMatch.home_game ? `Di Gröne - ${prevMatch.opponent}` : `${prevMatch.opponent} - Di Gröne`}
                  </p>
                  <p className={`text-2xl font-bold 
                    ${prevMatch.goals_for > prevMatch.goals_against ? 'text-green-600' : 
                    prevMatch.goals_for < prevMatch.goals_against ? 'text-red-600' : 
                    'text-grey-600'}`}>
                    {prevMatch.home_game ? `${prevMatch.goals_for} - ${prevMatch.goals_against}` : `${prevMatch.goals_against} - ${prevMatch.goals_for}`}
                  </p>
                </div>
              ) : (
                <p className="text-sm text-muted-foreground mt-4">Ingen tidigare match</p>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-0">
              <CardTitle className="font-medium">Kommande match</CardTitle>
            </CardHeader>
            <CardContent>
              {nextMatch ? (
                <div className="mt-4 text-center">
                  <p className="text-sm text-muted-foreground mb-2">
                    {new Date(nextMatch.start_time).toLocaleDateString('sv-SE', { 
                      day: 'numeric',
                      month: 'long',
                      year: 'numeric',
                      hour: '2-digit',
                      minute: '2-digit'
                    })}
                  </p>
                  <p className="text-lg">
                    {nextMatch.home_game ? `Di Gröne - ${nextMatch.opponent}` : `${nextMatch.opponent} - Di Gröne`}
                  </p>
                </div>
              ) : (
                <p className="text-sm text-muted-foreground mt-4">Ingen kommande match</p>
              )}
            </CardContent>
          </Card>
        </div>
      </main>
      <footer className="mt-16 py-6 border-t border-border/40">
        <div className="container mx-auto px-4 flex flex-col items-center justify-center">
          <p className="text-xs text-muted-foreground/60 mb-2">Framtagen och drivs av:</p>
          <img src="asklingforit.logo.svg" alt="Askling for IT" className="h-16" />
        </div>
      </footer>
    </div>
  );
};

export default Index;
