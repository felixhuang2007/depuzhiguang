class TableInfo {
  final String id;
  final String name;
  final String stakes;
  final int maxPlayers;
  final int currentPlayers;
  final int limit;
  final bool isFull;

  const TableInfo({
    required this.id,
    required this.name,
    required this.stakes,
    required this.maxPlayers,
    required this.currentPlayers,
    required this.limit,
    required this.isFull,
  });
}
