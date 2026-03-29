import 'package:flutter/material.dart';
import 'screens/status_screen.dart';
import 'screens/gogs_screen.dart';
import 'screens/feed_screen.dart';
import 'screens/db_screen.dart';

void main() {
  runApp(const BnsApp());
}

class BnsApp extends StatelessWidget {
  const BnsApp({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Bigwash Native Shell',
      theme: ThemeData(
        useMaterial3: true,
        brightness: Brightness.dark,
        scaffoldBackgroundColor: Colors.black,
        primaryColor: const Color(0xFF00FF41),
        colorScheme: ColorScheme.dark(
          primary: const Color(0xFF00FF41),
          secondary: const Color(0xFF00FF41),
          surface: Colors.grey[900] ?? Colors.grey,
        ),
        appBarTheme: const AppBarTheme(
          backgroundColor: Colors.black,
          foregroundColor: Color(0xFF00FF41),
        ),
        bottomNavigationBarTheme: BottomNavigationBarThemeData(
          backgroundColor: Colors.grey[900],
          selectedItemColor: const Color(0xFF00FF41),
          unselectedItemColor: Colors.grey,
        ),
      ),
      home: const BnsHome(),
    );
  }
}

class BnsHome extends StatefulWidget {
  const BnsHome({Key? key}) : super(key: key);

  @override
  State<BnsHome> createState() => _BnsHomeState();
}

class _BnsHomeState extends State<BnsHome> {
  int _currentIndex = 0;

  final List<Widget> _screens = const [
    StatusScreen(),
    GogsScreen(),
    FeedScreen(),
    DbScreen(),
  ];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: _screens[_currentIndex],
      bottomNavigationBar: BottomNavigationBar(
        currentIndex: _currentIndex,
        onTap: (index) {
          setState(() {
            _currentIndex = index;
          });
        },
        items: const [
          BottomNavigationBarItem(
            icon: Icon(Icons.show_chart),
            label: 'Status',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.source),
            label: 'Commits',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.signal_cellular_alt),
            label: 'Feed',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.database),
            label: 'Database',
          ),
        ],
      ),
    );
  }
}
